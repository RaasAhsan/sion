use std::io::{self, Cursor, Read, Seek, Write};

use self::{fs::FileSystem, response::ErrorData, storage::StorageClient};

pub mod fs;
mod metadata;
mod response;
mod storage;

const CHUNK_SIZE: usize = 8 * 1024 * 1024;

pub struct File {
    pub path: String,

    fs: FileSystem,
}

impl File {
    fn new(path: String, fs: FileSystem) -> File {
        File { path, fs }
    }
}

impl Read for File {
    fn read(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
        // std::fs::File::open(source_path).unwrap();
        todo!()
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
        let append_resp = self.fs.metadata.append_chunk(&self.path);
        match append_resp {
            Ok(append) => {
                // TODO: use real address here
                let storage = self.fs.connect_to_storage("http://localhost:8080");
                match storage.upload_chunk(append.chunk_id, buf) {
                    Ok(value) => Ok(value.received),
                    Err(_) => Err(io::Error::new(io::ErrorKind::Other, "upload failed")),
                }
            }
            Err(_) => Err(io::Error::new(io::ErrorKind::Other, "append failed")),
        }
    }

    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}

#[derive(Debug)]
pub enum Error {
    NetworkError,
    ResponseError,
    ServerError(ErrorData),
    Unknown,
}
