mod cli;
mod client;
mod util;

use crc32fast::Hasher;
use reqwest::blocking::{Body, Client};
use std::io::{self, Read};

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

fn main() {
    // upload_chunk();
    // upload_stream();
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
