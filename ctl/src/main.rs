use std::{io::{self, Read}, cmp, fmt::format, cell::RefCell, sync::{Arc, Mutex}};
use crc32fast::Hasher;
use reqwest::blocking::{Body, Client};

const BUFFER_SIZE: usize = 256;
const CHUNK_SIZE: usize = 8 * 1024 * 1024;

struct ChunkedReader<T: Read> {
    source: T,
    limit: usize,
    done: Arc<Mutex<bool>>
}

impl<T: Read> ChunkedReader<T> {
    fn new(source: T, limit: usize, done: Arc<Mutex<bool>>) -> ChunkedReader<T> {
        ChunkedReader { 
            source: source, 
            limit: limit,
            done: done
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
            },
            Ok(n) => self.limit -= n,
            // TODO: mark the reader as errored out?
            _ => {}
        }
        res
    }
}

fn upload_chunk() {
    let client = Client::new();
    let body = Body::new(io::stdin());
    let resp = client.post("http://localhost:8080/chunks/4")
        .body(body)
        .send()
        .unwrap();
    println!("{}", resp.text().unwrap());
}

fn upload_stream() {
    let mut id = 1;

    let client = Client::new();
    loop {
        let done = Arc::new(Mutex::new(false));
        let reader = ChunkedReader::new(io::stdin(), CHUNK_SIZE, done.clone());
        let body = Body::new(reader);
        let chunk_name = format!("chunk-{}", id);
        let resp = client.post(format!("http://localhost:8080/chunks/{}", chunk_name))
            .body(body)
            .send()
            .unwrap();

        println!("{}", resp.text().unwrap());

        if *done.lock().unwrap() {
            break;
        }
        id += 1;
    }
}

fn main() {
    // upload_chunk();
    upload_stream();
}

fn main2() {
    let mut buffer = [0; BUFFER_SIZE];
    let mut rb: usize = 0;

    let mut hasher = Hasher::new();

    let mut n: usize = usize::MAX;
    while n > 0 {
        // stdin is a buffered reader protected by a mutex;
        // if we want to be more efficient, we can lock and consume directly
        n = io::stdin().read(&mut buffer[..]).unwrap();
        rb += n;
        println!("{}", n);

        if n > 0 {
            hasher.update(&buffer[0..n]);
        }
    }

    let checksum = hasher.finalize();

    println!("Read {} bytes, checksum: {:x}", rb, checksum);
}
