use std::io::{Read, Seek};

use self::response::ErrorData;

pub mod fs;
mod metadata;
mod response;
mod storage;

pub struct File {
    pub path: String,
    pub size: u64,
    pub offset: u64,
}

impl File {
    fn new(path: String, size: u64) -> File {
        File {
            path,
            size,
            offset: 0,
        }
    }
}

impl Read for File {
    fn read(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
        todo!()
    }
}

impl Read for &File {
    fn read(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
        todo!()
    }
}

impl Seek for File {
    fn seek(&mut self, pos: std::io::SeekFrom) -> std::io::Result<u64> {
        todo!()
    }
}

impl Seek for &File {
    fn seek(&mut self, pos: std::io::SeekFrom) -> std::io::Result<u64> {
        todo!()
    }
}

#[derive(Debug)]
pub enum Error {
    NetworkError,
    ResponseError,
    ServerError(ErrorData),
    Unknown,
}
