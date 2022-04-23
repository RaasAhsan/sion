use reqwest::blocking::Client;

pub struct StorageClient {
    address: String,
    client: Client,
}

impl StorageClient {
    // TODO: reinitialize for each call or pass address in?
    pub fn new(address: String, client: Client) -> StorageClient {
        StorageClient { address, client }
    }

    fn upload_file(&self, chunk_id: &str) {}
}
