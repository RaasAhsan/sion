use reqwest::blocking::Client;

use std::{collections::HashMap, io};

trait MetadataClient {
    // Cluster operations
    fn get_cluster_mapping(&self) -> io::Result<HashMap<String, String>>;

    // Namespace operations
    fn get_file(&self, path: &str) -> io::Result<u32>;
    fn create_file(&self, path: &str) -> io::Result<u32>;
    fn append_chunk(&self, path: &str) -> io::Result<String>;
}

pub struct MetadataClientImpl {
    address: String,
    client: Client,
}

impl MetadataClientImpl {
    pub fn new(address: String, client: Client) -> MetadataClientImpl {
        MetadataClientImpl { address, client }
    }
}
