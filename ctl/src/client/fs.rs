use reqwest::blocking::{Body, Client};
use std::{
    io,
    sync::{Arc, Mutex},
};

use crate::util::chunked_reader::ChunkedReader;

use super::{File, metadata_client::MetadataClientImpl};

const CHUNK_SIZE: usize = 8 * 1024 * 1024;

// TODO: what is the interface we would like to expose here?
pub struct FileSystem {
    metadata_client: MetadataClientImpl
}

pub struct FileSystemImpl {}

impl FileSystem {

    fn connect(address: &str) -> FileSystem {

    }

    fn open(&self, path: &str) -> io::Result<File> {


        todo!()
    }
    // fn copy_stdin_to_remote(&self, dest_path: &str) -> io::Result<i32>;
    // fn copy_local_to_remote(&self, source_path: &str, dest_path: &str) -> io::Result<i32>;
    // fn copy_remote_to_local(&self, dest_path: &str, source_path: &str) -> io::Result<i32>;
    // fn copy_remote_to_remote(&self, dest_path: &str, source_path: &str) -> io::Result<i32>;

    fn copy_stdin_to_remote(&self, dest_path: &str) -> io::Result<i32> {
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

    fn copy_local_to_remote(&self, source_path: &str, dest_path: &str) -> io::Result<i32> {
        let mut id = 3;
        let file = std::fs::File::open(source_path).unwrap();

        let client = Client::new();
        loop {
            let done = Arc::new(Mutex::new(false));
            let reader = ChunkedReader::new(file, CHUNK_SIZE, done.clone());
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
