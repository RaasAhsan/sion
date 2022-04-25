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

    pub fn upload_chunk(&self, chunk_id: String, body: Body) -> Result<UploadChunkResponse, Error> {
        let resp = self
            .client
            .post(format!("{}/chunks/{}", self.address, chunk_id))
            .body(body)
            .send()
            .map_err(|_| Error::NetworkError)?;
        super::response::parse_from_response::<UploadChunkResponse>(resp)
    }
}

#[derive(Deserialize, Debug)]
pub struct UploadChunkResponse {
    #[serde(rename(deserialize = "Id"))]
    pub id: String,
    #[serde(rename(deserialize = "Received"))]
    pub received: usize,
}
