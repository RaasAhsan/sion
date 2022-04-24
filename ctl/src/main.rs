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

fn main4() {
    // upload_chunk();
    // upload_stream();

    let fs = client::fs::FileSystem::connect("http://localhost:8000").unwrap();

    let version = fs.metadata.version().unwrap();
    println!("{:?}", version);

    let mut file = fs.open("helloworld.txt").unwrap();

    let buf = vec![1, 2, 3, 4, 5];
    let bytes = file.write(&buf).unwrap();

    println!("wrote {} bytes", bytes);

    // let mapping = fs.metadata.get_cluster_mapping().unwrap();
    // println!("{:?}", mapping);

    // let file = fs.open("helloworld.txt").unwrap();
    // println!("{}", file.path);
}

fn main3() {
    let mut buffer = [0; BUFFER_SIZE];
    let wb = vec![1, 2, 3, 4, 5];

    let mut cursor: Cursor<Vec<u8>> = Cursor::new(Vec::new());
    let wr = cursor.write_all(&wb).unwrap();
    cursor.set_position(0);
    let result = cursor.read(&mut buffer).unwrap();

    println!("{}", result);
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

use std::sync::mpsc::{sync_channel, Receiver, SyncSender};
use std::thread;

enum Message {
    Chunk(Vec<u8>),
    Flush,
}

struct TxWrite {
    tx: SyncSender<Message>,
}

impl Write for TxWrite {
    fn write(&mut self, buf: &[u8]) -> io::Result<usize> {
        // TODO: can we avoid copying?
        self.tx.send(Message::Chunk(buf.to_vec())).unwrap(); // TODO: handle error
        Ok(buf.len())
    }

    fn flush(&mut self) -> io::Result<()> {
        self.tx.send(Message::Flush).unwrap(); // TODO: handle error
        Ok(())
    }
}

struct TxRead {
    rx: Receiver<Message>,
    // Technically we are buffering up to double a single message
    partial: Option<Vec<u8>>,
}

impl Read for TxRead {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        let incoming = match self.partial.take() {
            Some(bytes) => Some(bytes),
            None => {
                let message = self.rx.recv().unwrap();
                match message {
                    Message::Chunk(bytes) => Some(bytes),
                    Message::Flush => None,
                }
            }
        };
        match incoming {
            Some(bytes) => {
                let len = std::cmp::min(buf.len(), bytes.len());
                &buf[..len].clone_from_slice(&bytes[..len]);

                if bytes.len() <= buf.len() {
                    self.partial = None;
                } else {
                    self.partial = Some((&bytes[len..]).to_vec());
                }

                Ok(len)
            }
            None => Ok(0),
        }
    }
}

fn main() {
    let (tx, rx) = sync_channel::<Message>(0);

    let tx_writer = TxWrite { tx };
    let mut tx_reader = TxRead { rx, partial: None };
    let mut writer = BufWriter::new(tx_writer);

    let handle = thread::spawn(move || {
        let mut buf = Vec::new();
        tx_reader.read_to_end(&mut buf);
        println!("{:?}", buf);
    });

    writer.write(&[1, 2, 3, 4, 5]).unwrap();
    writer.write(&[1, 2, 3, 4, 5]).unwrap();
    writer.flush();

    handle.join();
    println!("joined!");
}
