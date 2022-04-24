use reqwest::{
    blocking::{Body, Client},
    StatusCode,
};
use serde::{Deserialize, Serialize};

use std::{collections::HashMap, io};

use super::Error;

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

    pub fn get_cluster_mapping(&self) -> Result<GetClusterMappingResponse, ()> {
        let resp = self
            .client
            .get(format!("{}/nodes", self.address))
            .send()
            .map_err(|_| ())?;

        match resp.status() {
            StatusCode::OK => {
                let get_mapping: GetClusterMappingResponse = serde_json::from_reader(resp).map_err(|_| ())?;
                Ok(get_mapping)
            }
            _ => Result::Err(()),
        }
    }

    // Namespace operations

    pub fn get_file(&self, path: &str) -> Result<FileResponse, Error> {
        let resp = self
            .client
            .get(format!("{}/files/{}", self.address, path))
            .send()
            .map_err(|_| Error::NetworkError)?;

        match resp.status() {
            StatusCode::OK => {
                let get_file: FileResponse = serde_json::from_reader(resp).map_err(|_| Error::ResponseError)?;
                Ok(get_file)
            },
            StatusCode::NOT_FOUND => Err(Error::FileNotFound),
            _ => Err(Error::Unknown),
        }
    }

    pub fn create_file(&self, path: &str) -> Result<FileResponse, ()> {
        let resp = self
            .client
            .post(format!("{}/files/{}", self.address, path))
            .send()
            .map_err(|_| ())?;

        match resp.status() {
            StatusCode::OK => {
                let get_file: FileResponse = serde_json::from_reader(resp).map_err(|_| ())?;
                Ok(get_file)
            }
            _ => Result::Err(()),
        }
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

#[derive(Deserialize, Debug)]
pub struct GetClusterMappingResponse {
    #[serde(rename(deserialize = "Addresses"))]
    pub addresses: HashMap<String, String>
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
