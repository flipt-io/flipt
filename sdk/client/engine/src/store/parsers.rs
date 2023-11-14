use snafu::{prelude::*, Whatever};
use std::env;

#[cfg(test)]
use std::fs;
#[cfg(test)]
use std::path::PathBuf;

use super::snapshot::Parser;
use crate::models::transport;

pub struct FliptParser {
    namespaces: Vec<String>,
    http_client: reqwest::blocking::Client,
    http_url: String,
}

impl FliptParser {
    pub fn new(namespaces: Vec<String>) -> Self {
        // We will allow the following line to panic when an error is encountered.
        let http_client = reqwest::blocking::Client::builder()
            .timeout(std::time::Duration::from_secs(10))
            .build()
            .unwrap();

        let http_url =
            env::var("FLIPT_REMOTE_URL").unwrap_or(String::from("http://localhost:8080"));

        Self {
            namespaces,
            http_client,
            http_url,
        }
    }
}

impl Parser for FliptParser {
    fn parse(&self, namespace: String) -> Result<transport::Document, Whatever> {
        let response = match self
            .http_client
            .get(format!(
                "{}/internal/v1/evaluation/snapshot/namespace/{}",
                self.http_url, namespace
            ))
            .send()
        {
            Ok(resp) => resp,
            Err(e) => whatever!("failed to make request: err {}", e),
        };

        let response_text = match response.text() {
            Ok(t) => t,
            Err(e) => whatever!("failed to get response body: err {}", e),
        };

        let document: transport::Document = match serde_json::from_str(&response_text) {
            Ok(doc) => doc,
            Err(e) => whatever!("failed to deserialize text into document: err {}", e),
        };

        Ok(document)
    }

    fn get_namespaces(&self) -> Vec<String> {
        self.namespaces.clone()
    }
}

#[cfg(test)]
pub struct TestParser {
    namespaces: Vec<String>,
}

#[cfg(test)]
impl TestParser {
    pub fn new(namespaces: Vec<String>) -> Self {
        Self { namespaces }
    }
}

#[cfg(test)]
impl Parser for TestParser {
    fn parse(&self, _: String) -> Result<transport::Document, Whatever> {
        let mut d = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
        d.push("src/testdata/state.json");

        let state =
            fs::read_to_string(d.display().to_string()).expect("file should have read correctly");

        let document: transport::Document = match serde_json::from_str(&state) {
            Ok(document) => document,
            Err(e) => whatever!("failed to deserialize text into document: err {}", e),
        };

        Ok(document)
    }

    fn get_namespaces(&self) -> Vec<String> {
        self.namespaces.clone()
    }
}
