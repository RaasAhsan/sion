mod cli;
mod client;
mod util;

use crc32fast::Hasher;
use reqwest::blocking::{Body, Client};
use std::{
    io::{self, BufWriter, Cursor, Read, Write},
    net::TcpStream,
};

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
    let fs = client::fs::FileSystem::connect("http://localhost:8000").unwrap();

    let version = fs.metadata.version().unwrap();
    println!("{:?}", version);

    let mut file = fs.open("helloworld.txt").unwrap();

    // let buf = vec![1, 2, 3, 4, 5];
    // let bytes = file.write(&buf).unwrap();
    // let bytes2 = file.write(&buf).unwrap();

    // std::thread::sleep(std::time::Duration::from_millis(5000));

    // file.flush().unwrap();

    // println!("wrote {} bytes", bytes + bytes2);

    let chunks = fs.metadata.get_chunks(file.path.as_str()).unwrap();

    let storage = fs.connect_to_storage("http://localhost:8080");

    let mut cursor = Cursor::new(Vec::new());

    for chunk in chunks {
        println!("{:?}", chunk);
        storage.download_chunk(chunk.chunk_id, &mut cursor).unwrap();
    }

    let bytes = &cursor.get_ref()[..];
    println!("{:?}", bytes);

    // let mapping = fs.metadata.get_cluster_mapping().unwrap();
    // println!("{:?}", mapping);

    // let file = fs.open("helloworld.txt").unwrap();
    // println!("{}", file.path);
}

struct Conn {
    stream: TcpStream,
}

fn test(conn: &mut Conn) {
    let mut buf = vec![1, 2, 3, 4, 5];
    let stream = &conn.stream;
    conn.stream.write(&buf);
    conn.stream.write(&buf);
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
