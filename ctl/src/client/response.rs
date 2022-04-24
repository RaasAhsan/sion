use serde::{de::DeserializeOwned, Deserialize};

use super::Error;

#[derive(Deserialize, Debug)]
pub enum Response<T> {
    Success(T),
    Error(ErrorData),
}

#[derive(Deserialize, Debug)]
pub struct ErrorData {
    #[serde(rename(deserialize = "Message"))]
    pub message: String,
    #[serde(rename(deserialize = "Code"))]
    pub code: ErrorCode,
}

#[derive(Deserialize, Debug)]
pub enum ErrorCode {
    FileNotFound,
    ChunkNotFound,
    Unknown,
}

pub fn parse_from_response<T: DeserializeOwned>(
    resp: reqwest::blocking::Response,
) -> Result<T, Error> {
    let parsed: Response<T> = serde_json::from_reader(resp).map_err(|_| Error::ResponseError)?;
    match parsed {
        Response::Success(value) => Ok(value),
        Response::Error(e) => Err(Error::ServerError(e)),
    }
}
