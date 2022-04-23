use reqwest::{
    blocking::{Body, Client},
    StatusCode,
};
use serde::{Deserialize, Serialize};

use std::{collections::HashMap, io};

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

    fn get_cluster_mapping(&self) -> io::Result<HashMap<String, String>> {
        todo!()
    }

    // Namespace operations

    pub fn get_file(&self, path: &str) -> Result<GetFileResponse, ()> {
        let resp = self
            .client
            .get(format!("{}/files/{}", self.address, path))
            .send()
            .map_err(|_| ())?;

        match resp.status() {
            StatusCode::OK => {
                let get_file: GetFileResponse = serde_json::from_reader(resp).map_err(|_| ())?;
                Ok(get_file)
            }
            _ => Result::Err(()),
        }
    }

    fn create_file(&self, path: &str) -> io::Result<u32> {
        todo!()
    }

    fn append_chunk(&self, path: &str) -> io::Result<String> {
        todo!()
    }

    pub fn version(&self) -> Result<VersionResponse, ()> {
        // let body = Body::new(io::stdin());
        let resp = self
            .client
            .get(format!("{}/version", self.address))
            .send()
            .map_err(|_| ())?;

        match resp.status() {
            StatusCode::OK => {
                let version: VersionResponse = serde_json::from_reader(resp).map_err(|_| ())?;
                Ok(version)
            }
            _ => Result::Err(()),
        }
    }
}

#[derive(Deserialize)]
pub struct GetClusterMappingResponse {

}

#[derive(Deserialize)]
pub struct GetFileResponse {
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
