use std::{
    io::{self, BufReader, BufWriter, Read, Seek, Write},
    ops::Index,
    thread::{self, JoinHandle},
};

use http_content_range::ContentRange;
use reqwest::blocking::Body;

use self::{buffer::TxWrite, fs::FileSystem, response::ErrorData};

mod buffer;
pub mod fs;
mod metadata;
mod response;
mod storage;

pub struct File {
    pub path: String,

    fs: FileSystem,
    read_state: Option<ReadState>,
    write_state: Option<WriteState>,
}

struct ReadState {
    chunks: Vec<String>,
    chunk_index: usize,
    chunk_offset: usize,
}

struct WriteState {
    handle: JoinHandle<std::io::Result<usize>>, // TODO: return result of response
    sender: BufWriter<TxWrite>,
}

impl File {
    fn new(path: String, fs: FileSystem) -> File {
        File {
            path,
            fs,
            read_state: None,
            write_state: None,
        }
    }

    // For now, assume the buffer is always large enough to read a whole chunk
    // TODO: fix this once we have more size information
    fn read_one_chunk(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
        let state = self.read_state.as_mut().unwrap();
        if state.chunk_index >= state.chunks.len() {
            Ok(0)
        } else {
            let storage = self.fs.connect_to_storage("http://localhost:8080");

            let start = state.chunk_offset;
            // It is fine if we go past the end of the file; server will tell us the true length
            // TODO: but fix this later when we have chunk metadata locally
            let end = start + buf.len() - 1;
            let mut vbuf = Vec::new();
            match storage.download_chunk(
                state.chunks.index(state.chunk_index),
                &mut vbuf,
                Some((start, end)),
            ) {
                Ok(resp) => {
                    buf[..vbuf.len()].clone_from_slice(&vbuf);

                    // Check chunk boundary to see if we need to request more
                    // bytes on the current chunk, or if we can move onto the next
                    match chunk_boundary_from_range(&resp.content_range) {
                        ChunkBoundary::Next => {
                            state.chunk_index += 1;
                            state.chunk_offset = 0;
                            Ok(resp.bytes)
                        }
                        ChunkBoundary::More(offset) => {
                            state.chunk_offset = offset;
                            Ok(resp.bytes)
                        }
                        ChunkBoundary::Unknown => {
                            Err(io::Error::new(io::ErrorKind::Other, "invalid chunk range"))
                        }
                    }
                }
                Err(_) => Err(io::Error::new(
                    io::ErrorKind::Other,
                    "failed to download chunk",
                )),
            }
        }
    }
}

impl Read for File {
    fn read(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
        if let Some(_) = &mut self.read_state {
            self.read_one_chunk(buf)
        } else {
            let chunks_resp = self.fs.metadata.get_chunks(&self.path);
            match chunks_resp {
                Ok(resp) => {
                    let chunks: Vec<String> = resp.into_iter().map(|r| r.chunk_id).collect();
                    self.read_state = Some(ReadState {
                        chunks,
                        chunk_index: 0,
                        chunk_offset: 0,
                    });
                    self.read_one_chunk(buf)
                }
                Err(_) => Err(io::Error::new(io::ErrorKind::Other, "failed to get chunks")),
            }
        }
    }
}

impl Seek for File {
    fn seek(&mut self, pos: std::io::SeekFrom) -> std::io::Result<u64> {
        todo!()
    }
}

impl Write for File {
    // We will automatically buffer files.
    // Could potentially offer an unbuffered version which calls append every time.
    // Chunks are flushed and committed when flush is called.

    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        if let Some(state) = &mut self.write_state {
            state.sender.write(buf)
        } else {
            let append_resp = self.fs.metadata.append_chunk(&self.path);
            match append_resp {
                Ok(append) => {
                    let (tx_writer, reader) = buffer::channel();
                    let mut writer = BufWriter::new(tx_writer);

                    let storage = self.fs.connect_to_storage("http://localhost:8080");

                    let handle = thread::spawn(move || {
                        // TODO: use real address here
                        match storage.upload_chunk(&append.chunk_id, Body::new(reader)) {
                            Ok(value) => Ok(value.received),
                            Err(_) => Err(io::Error::new(io::ErrorKind::Other, "upload failed")),
                        }
                    });
                    let init_write = writer.write(buf);
                    self.write_state = Some(WriteState {
                        handle,
                        sender: writer,
                    });
                    init_write
                }
                Err(_) => Err(io::Error::new(io::ErrorKind::Other, "append failed")),
            }
        }
    }

    // fn write_unbuffered(&mut self, buf: &[u8]) -> io::Result<usize> {
    //     let append_resp = self.fs.metadata.append_chunk(&self.path);
    //     match append_resp {
    //         Ok(append) => {
    //             // TODO: use real address here
    //             let storage = self.fs.connect_to_storage("http://localhost:8080");
    //             match storage.upload_chunk(append.chunk_id, buf.to_vec().into()) {
    //                 Ok(value) => Ok(value.received),
    //                 Err(_) => Err(io::Error::new(io::ErrorKind::Other, "upload failed")),
    //             }
    //         }
    //         Err(_) => Err(io::Error::new(io::ErrorKind::Other, "append failed")),
    //     }
    // }

    fn flush(&mut self) -> io::Result<()> {
        match self.write_state.take() {
            Some(mut state) => {
                let res = state.sender.flush();
                state.handle.join().unwrap(); // TODO: incorporate this in the flush
                res
            }
            None => Ok(()),
        }
    }
}

#[derive(Debug)]
pub enum Error {
    NetworkError,
    ResponseError,
    ServerError(ErrorData),
    Unknown,
}

#[derive(Debug)]
enum ChunkBoundary {
    Next,
    More(usize),
    Unknown,
}

fn chunk_boundary_from_range(range: &Option<ContentRange>) -> ChunkBoundary {
    match range {
        Some(range) => match range {
            ContentRange::Bytes(range) => {
                if range.last_byte + 1 == range.complete_length {
                    ChunkBoundary::Next
                } else {
                    ChunkBoundary::More((range.last_byte + 1) as usize)
                }
            }
            ContentRange::UnboundBytes(_) => ChunkBoundary::Next,
            ContentRange::Unsatisfied(_) => ChunkBoundary::Unknown,
            ContentRange::Unknown => ChunkBoundary::Unknown,
        },
        None => ChunkBoundary::Next,
    }
}
