use reqwest::blocking::Client;
use serde::Deserialize;

use std::collections::HashMap;

use super::Error;

#[derive(Clone)]
pub struct MetadataClient {
    address: String,
    client: Client,
}

impl MetadataClient {
    pub fn new(address: &str, client: Client) -> MetadataClient {
        MetadataClient {
            address: String::from(address),
            client,
        }
    }

    // Cluster operations

    pub fn get_cluster_mapping(&self) -> Result<GetClusterMappingResponse, Error> {
        self.client
            .get(format!("{}/nodes", self.address))
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|resp| super::response::parse_from_response(resp))
    }

    // Namespace operations

    pub fn get_file(&self, path: &str) -> Result<FileResponse, Error> {
        self.client
            .get(format!("{}/files/{}", self.address, path))
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|resp| super::response::parse_from_response(resp))
    }

    pub fn create_file(&self, path: &str) -> Result<FileResponse, Error> {
        self.client
            .post(format!("{}/files/{}", self.address, path))
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|resp| super::response::parse_from_response(resp))
    }

    pub fn get_chunks(&self, path: &str) -> Result<Vec<ChunkLocation>, Error> {
        self.client
            .get(format!("{}/files/{}/chunks", self.address, path))
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|resp| super::response::parse_from_response(resp))
    }

    // TODO: append to old chunk
    pub fn freeze_chunk(&self, path: &str, chunk_id: &str) -> Result<ChunkLocation, Error> {
        self.client
            .post(format!("{}/files/{}/chunks/{}/freeze", self.address, path, chunk_id))
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|resp| super::response::parse_from_response(resp))
    }

    pub fn version(&self) -> Result<VersionResponse, Error> {
        // let body = Body::new(io::stdin());
        self.client
            .get(format!("{}/version", self.address))
            .send()
            .map_err(|_| Error::NetworkError)
            .and_then(|resp| super::response::parse_from_response(resp))
    }
}

#[derive(Deserialize, Debug)]
pub struct GetClusterMappingResponse {
    #[serde(rename(deserialize = "Addresses"))]
    pub addresses: HashMap<String, String>,
}

#[derive(Deserialize, Debug)]
pub struct FileResponse {
    #[serde(rename(deserialize = "Path"))]
    pub path: String,
    #[serde(rename(deserialize = "TimeCreated"))]
    pub time_created: u64,
    #[serde(rename(deserialize = "TimeModified"))]
    pub time_modified: u64,
    #[serde(rename(deserialize = "Size"))]
    pub size: u64,
    #[serde(rename(deserialize = "TailChunk"))]
    pub tail_chunk: ChunkLocation,
}

#[derive(Deserialize, Debug)]
pub struct VersionResponse {
    #[serde(rename(deserialize = "ServerVersion"))]
    pub server_version: String,
    #[serde(rename(deserialize = "MajorVersion"))]
    pub major_version: u8,
    #[serde(rename(deserialize = "MinorVersion"))]
    pub minor_version: u8,
    #[serde(rename(deserialize = "PatchVersion"))]
    pub patch_version: u8,
}

#[derive(Deserialize, Debug)]
pub struct ChunkLocation {
    #[serde(rename(deserialize = "Id"))]
    pub chunk_id: String,
    #[serde(rename(deserialize = "Nodes"))]
    pub node_ids: Vec<String>,
}
