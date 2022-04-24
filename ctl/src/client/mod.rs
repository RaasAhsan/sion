use std::{io::{Read, Seek, Write, Cursor, self}, thread::{self, JoinHandle}};

use reqwest::blocking::Body;

use self::{fs::FileSystem, response::ErrorData};

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
        File {
            path,

            fs,
        }
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

    fn write(&mut self, buf: &[u8]) -> std::io::Result<usize> {
        // let resp = client
        //     .post(format!("http://localhost:8080/chunks/{}", chunk_name))
        //     .body(body)
        //     .send()
        //     .unwrap();

        todo!()
    }

    fn flush(&mut self) -> io::Result<()> {
        io::Result::Ok(())
    }
}

#[derive(Debug)]
pub enum Error {
    NetworkError,
    ResponseError,
    ServerError(ErrorData),
    Unknown,
}
