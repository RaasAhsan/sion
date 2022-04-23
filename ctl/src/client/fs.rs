use reqwest::blocking::{Body, Client};
use std::{
    io,
    sync::{Arc, Mutex},
};

use crate::util::chunked_reader::ChunkedReader;

use super::{metadata::MetadataClient, File};

const CHUNK_SIZE: usize = 8 * 1024 * 1024;

// TODO: what is the interface we would like to expose here?
pub struct FileSystem {
    pub metadata: MetadataClient,
    client: Client,
}

impl FileSystem {
    // TODO: check version and fail if incompatible
    pub fn connect(address: &str) -> FileSystem {
        let client = Client::new();
        let metadata = MetadataClient::new(address, client.clone());
        FileSystem { metadata, client }
    }

    pub fn open(&self, path: &str) -> Result<File, ()> {
        let get_file = self.metadata.get_file(path)?;
        let file = File::new(get_file.path, get_file.size);
        Ok(file)
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

    // fn copy_local_to_remote(&self, source_path: &str, dest_path: &str) -> io::Result<i32> {
    //     let mut id = 3;
    //     let file = std::fs::File::open(source_path).unwrap();

    //     let client = Client::new();
    //     loop {
    //         let done = Arc::new(Mutex::new(false));
    //         let reader = ChunkedReader::new(&file, CHUNK_SIZE, done.clone());
    //         let body = Body::new(reader);
    //         let chunk_name = format!("chunk-{}", id);
    //         let resp = client
    //             .post(format!("http://localhost:8080/chunks/{}", chunk_name))
    //             .body(body)
    //             .send()
    //             .unwrap();

    //         println!("{}", resp.text().unwrap());

    //         if *done.lock().unwrap() {
    //             break;
    //         }
    //         id += 1;
    //     }

    //     Result::Ok(2)
    // }
}
