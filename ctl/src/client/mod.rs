use std::io::{Read, Seek};

pub mod fs;
mod metadata;
mod storage;

pub struct File {
    path: String,
    size: u64,
    offset: u64
}

impl File {
    fn new(path: String, size: u64) -> File {
        File { path, size, offset: 0 }
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
