mod chunked_reader;
mod metadata_client;
mod storage_client;

use crc32fast::Hasher;
use reqwest::blocking::{Body, Client};
use storage_client::StorageClient;
use std::{
    io::{self, Read, Result},
    sync::{Arc, Mutex}, fs::File,
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

trait FileSystem {
    fn copy_stdin_to_remote(&self, dest_path: &str) -> Result<i32>;
    fn copy_local_to_remote(&self, source_path: &str, dest_path: &str) -> Result<i32>;
}

// implement: cat, cp

pub struct FileSystemImpl {
    
}

impl FileSystem for FileSystemImpl {
    fn copy_stdin_to_remote(&self, dest_path: &str) -> Result<i32> {
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

        Result::Ok(2)
    }

    fn copy_local_to_remote(&self, source_path: &str, dest_path: &str) -> Result<i32> {
        let mut id = 3;
        let file = File::open(source_path).unwrap();
    
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

        Result::Ok(2)
    }
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
