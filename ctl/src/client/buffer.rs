use std::io;
use std::io::{Read, Write};
use std::sync::mpsc::{sync_channel, Receiver, SyncSender};

enum Message {
    Chunk(Vec<u8>),
    Flush,
}

pub fn channel() -> (TxWrite, TxRead) {
    let (tx, rx) = sync_channel(0);
    (TxWrite { tx }, TxRead { rx, partial: None })
}

pub struct TxWrite {
    tx: SyncSender<Message>,
}

// TODO: keep track of how many bytes have been sent?

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

pub struct TxRead {
    rx: Receiver<Message>,
    // Technically we are buffering up to double a single message
    partial: Option<Vec<u8>>,
}

impl Read for TxRead {
    fn read(&mut self, buf: &mut [u8]) -> io::Result<usize> {
        // TODO: potentially use peek to avoid double buffering?

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

                if bytes.len() > buf.len() {
                    self.partial = Some((&bytes[len..]).to_vec());
                }

                Ok(len)
            }
            None => Ok(0),
        }
    }
}

// fn main() {
//     let (tx, rx) = sync_channel::<Message>(0);

//     let tx_writer = TxWrite { tx };
//     let mut tx_reader = TxRead { rx, partial: None };
//     let mut writer = BufWriter::new(tx_writer);

//     let handle = thread::spawn(move || {
//         // Should this thread be responsible for just mkaing requests?
//         // Could have write, flush, close, etc.
//         // thread per request or thread per file lifetime
//         loop {
//             let mut buf = Vec::new();
//             tx_reader.read_to_end(&mut buf).unwrap();
//             println!("{:?}", buf);
//         }
//     });

//     // calling grammar: write* flush
//     writer.write(&[1, 2, 3, 4, 5]).unwrap();
//     writer.write(&[1, 2, 3, 4, 5]).unwrap();
//     writer.flush().unwrap();

//     writer.write(&[1, 2, 3]).unwrap();
//     writer.write(&[1, 2, 3]).unwrap();
//     writer.flush().unwrap();

//     handle.join().unwrap();
//     println!("joined!");
// }
