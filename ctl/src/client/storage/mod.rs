use std::io::Write;

use reqwest::blocking::{Body, Client};
use serde::Deserialize;

use super::Error;

pub struct StorageClient {
    address: String,
    client: Client,
}

impl StorageClient {
    // TODO: reinitialize for each call or pass address in?
    pub fn new(address: String, client: Client) -> StorageClient {
        StorageClient { address, client }
    }

    pub fn download_chunk<W: Write>(&self, chunk_id: String, writer: &mut W) -> Result<u64, Error> {
        self
            .client
            .get(format!("{}/chunks/{}", self.address, chunk_id))
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|mut resp| resp.copy_to(writer).map_err(|_| Error::NetworkError))
    }

    pub fn upload_chunk(&self, chunk_id: String, body: Body) -> Result<UploadChunkResponse, Error> {
        self.client
            .post(format!("{}/chunks/{}", self.address, chunk_id))
            .body(body)
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|resp| super::response::parse_from_response(resp))
    }
}

#[derive(Deserialize, Debug)]
pub struct UploadChunkResponse {
    #[serde(rename(deserialize = "Id"))]
    pub id: String,
    #[serde(rename(deserialize = "Received"))]
    pub received: usize,
}
