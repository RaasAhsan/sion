use std::io::{self, Read};
use crc32fast::Hasher;

const BUFFER_SIZE: usize = 256;
const BLOCK_SIZE: usize = 8 * 1024 * 1024;

fn main() {
    let mut buffer = [0; BUFFER_SIZE];
    let mut n: usize = usize::MAX;
    let mut rb: usize = 0;

    let mut hasher = Hasher::new();

    while n > 0 {
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
