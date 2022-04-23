use reqwest::{blocking::{Client, Body}, StatusCode};
use serde::{Serialize, Deserialize};

use std::{collections::HashMap, io};

pub struct MetadataClient {
    address: String,
    client: Client,
}

impl MetadataClient {

    pub fn new(address: &str, client: Client) -> MetadataClient {
        MetadataClient { address: String::from(address), client }
    }

    // Cluster operations

    fn get_cluster_mapping(&self) -> io::Result<HashMap<String, String>> {
        todo!()
    }

    // Namespace operations

    fn get_file(&self, path: &str) -> io::Result<u32>  {
        todo!()
    }

    fn create_file(&self, path: &str) -> io::Result<u32>  {
        todo!()
    }

    fn append_chunk(&self, path: &str) -> io::Result<String>  {
        todo!()
    }

    pub fn version(&self) -> Result<String, ()> {
        // let body = Body::new(io::stdin());
        let resp = self.client
            .get(format!("{}/version", self.address))
            .send()
            .map_err(|_| ())?;
        
        match resp.status() {
            StatusCode::OK => {
                let version: VersionResponse = serde_json::from_reader(resp).map_err(|_| ())?;
                Ok(version.server_version)
            },
            _ => Result::Err(())
        }
    }
}

#[derive(Deserialize)]
pub struct VersionResponse {
    #[serde(rename(deserialize = "ServerVersion"))]
    server_version: String
}
