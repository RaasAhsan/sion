use serde::Deserialize;

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
    Unknown
}
