use snafu::{prelude::*, Whatever};
use std::env;
use std::fs;
use std::path::PathBuf;

use super::models;
use super::snapshot::Parser;

pub struct FliptParser {
    namespaces: Vec<String>,
    http_client: reqwest::blocking::Client,
    http_url: String,
}

// TODO(yquansah): Implement network request here for fetching flag state from upstream
impl Parser for FliptParser {
    fn new(namespaces: Vec<String>) -> Self {
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

    fn parse(&self, namespace: String) -> Result<models::Document, Whatever> {
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

        let document: models::Document = match serde_json::from_str(&response_text) {
            Ok(doc) => doc,
            Err(e) => whatever!("failed to deserialize text into document: err {}", e),
        };

        Ok(document)
    }

    fn get_namespaces(&self) -> Vec<String> {
        self.namespaces.clone()
    }
}

pub struct TestParser {
    namespaces: Vec<String>,
}

impl Parser for TestParser {
    fn new(namespaces: Vec<String>) -> Self {
        Self { namespaces }
    }

    fn parse(&self, _: String) -> Result<models::Document, Whatever> {
        let mut d = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
        d.push("src/testdata/state.json");

        let state =
            fs::read_to_string(d.display().to_string()).expect("file should have read correctly");

        let document: models::Document = match serde_json::from_str(&state) {
            Ok(document) => document,
            Err(e) => whatever!("failed to deserialize text into document: err {}", e),
        };

        Ok(document)
    }

    fn get_namespaces(&self) -> Vec<String> {
        self.namespaces.clone()
    }
}
