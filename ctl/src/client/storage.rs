use std::io::Write;

use http_content_range::ContentRange;
use reqwest::{
    blocking::{Body, Client},
    header::{HeaderMap, HeaderValue, CONTENT_RANGE, RANGE},
    StatusCode,
};
use serde::Deserialize;

use super::Error;

pub struct StorageClient {
    address: String,
    client: Client,
}

impl StorageClient {
    // TODO: reinitialize for each call or pass address in?
    pub fn new(address: &str, client: Client) -> StorageClient {
        StorageClient { address: address.to_string(), client }
    }

    pub fn download_chunk<W: Write>(
        &self,
        chunk_id: &str,
        writer: &mut W,
        range: Option<(usize, usize)>,
    ) -> Result<DownloadChunkResponse, Error> {
        let mut headers = HeaderMap::new();
        if let Some((start, end)) = range {
            headers.insert(
                RANGE,
                HeaderValue::from_str(&format!("bytes={}-{}", start, end)).unwrap(),
            );
        }

        self.client
            .get(format!("{}/chunks/{}", self.address, chunk_id))
            .headers(headers)
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|mut resp| match resp.status() {
                StatusCode::OK | StatusCode::PARTIAL_CONTENT => resp
                    .copy_to(writer)
                    .map(|bytes| {
                        let content_range = if let Some(value) = resp.headers().get(CONTENT_RANGE) {
                            Some(ContentRange::parse_bytes(value.as_bytes()))
                        } else {
                            None
                        };
                        DownloadChunkResponse {
                            bytes: bytes as usize,
                            content_range: content_range,
                        }
                    })
                    .map_err(|_| Error::NetworkError),
                _ => Err(Error::ResponseError),
            })
    }

    pub fn upload_chunk(&self, chunk_id: &str, body: Body) -> Result<UploadChunkResponse, Error> {
        self.client
            .post(format!("{}/chunks/{}", self.address, chunk_id))
            .body(body)
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|resp| super::response::parse_from_response(resp))
    }

    pub fn append_chunk(&self, chunk_id: &str, body: Body) -> Result<AppendChunkResponse, Error> {
        self.client
            .patch(format!("{}/chunks/{}", self.address, chunk_id))
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

#[derive(Deserialize, Debug)]
pub struct AppendChunkResponse {
    #[serde(rename(deserialize = "Offset"))]
    pub offset: usize,
    #[serde(rename(deserialize = "Length"))]
    pub length: usize,
}

#[derive(Debug)]
pub struct DownloadChunkResponse {
    pub bytes: usize,
    pub content_range: Option<ContentRange>,
}
