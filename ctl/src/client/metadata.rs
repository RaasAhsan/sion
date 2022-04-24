use reqwest::blocking::{Body, Client};
use serde::{de::DeserializeOwned, Deserialize, Serialize};

use std::{collections::HashMap, io};

use super::{response::Response, Error};

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
        let resp = self
            .client
            .get(format!("{}/nodes", self.address))
            .send()
            .map_err(|_| Error::NetworkError)?;
        super::response::parse_from_response::<GetClusterMappingResponse>(resp)
    }

    // Namespace operations

    pub fn get_file(&self, path: &str) -> Result<FileResponse, Error> {
        let resp = self
            .client
            .get(format!("{}/files/{}", self.address, path))
            .send()
            .map_err(|_| Error::NetworkError)?;
        super::response::parse_from_response::<FileResponse>(resp)
    }

    pub fn create_file(&self, path: &str) -> Result<FileResponse, Error> {
        let resp = self
            .client
            .post(format!("{}/files/{}", self.address, path))
            .send()
            .map_err(|_| Error::NetworkError)?;
        super::response::parse_from_response::<FileResponse>(resp)
    }

    pub fn append_chunk(&self, path: &str) -> Result<AppendChunkResponse, Error> {
        let resp = self
            .client
            .post(format!("{}/files/{}/chunks", self.address, path))
            .send()
            .map_err(|_| Error::NetworkError)?;
        super::response::parse_from_response::<AppendChunkResponse>(resp)
    }

    pub fn version(&self) -> Result<VersionResponse, Error> {
        // let body = Body::new(io::stdin());
        let resp = self
            .client
            .get(format!("{}/version", self.address))
            .send()
            .map_err(|_| Error::NetworkError)?;
        super::response::parse_from_response::<VersionResponse>(resp)
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
pub struct AppendChunkResponse {
    #[serde(rename(deserialize = "ChunkId"))]
    pub chunk_id: String,
    #[serde(rename(deserialize = "NodeId"))]
    pub node_id: String,
}
