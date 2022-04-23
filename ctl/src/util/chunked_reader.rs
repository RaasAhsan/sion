use std::{
    cmp,
    io::{self, Read},
    sync::{Arc, Mutex},
};

pub struct ChunkedReader<T: Read> {
    source: T,
    limit: usize,
    done: Arc<Mutex<bool>>,
}

impl<T: Read> ChunkedReader<T> {
    pub fn new(source: T, limit: usize, done: Arc<Mutex<bool>>) -> ChunkedReader<T> {
        ChunkedReader {
            source: source,
            limit: limit,
            done: done,
        }
    }
}

impl<T: Read> Read for ChunkedReader<T> {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        if self.limit == 0 {
            return Ok(0);
        }

        let len = cmp::min(self.limit, buf.len());
        let res = self.source.read(&mut buf[..len]);
        match res {
            Ok(0) => {
                let mut done = self.done.lock().unwrap();
                *done = true;
                self.limit = 0;
            }
            Ok(n) => self.limit -= n,
            // TODO: mark the reader as errored out?
            _ => {}
        }
        res
    }
}
