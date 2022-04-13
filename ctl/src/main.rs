mod chunked_reader;

use crc32fast::Hasher;
use reqwest::blocking::{Body, Client};
use std::{
    io::{self, Read},
    sync::{Arc, Mutex},
};

use chunked_reader::ChunkedReader;

const BUFFER_SIZE: usize = 256;
const CHUNK_SIZE: usize = 8 * 1024 * 1024;

fn upload_chunk() {
    let client = Client::new();
    let body = Body::new(io::stdin());
    let resp = client
        .post("http://localhost:8080/chunks/4")
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
        let resp = client
            .post(format!("http://localhost:8080/chunks/{}", chunk_name))
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
