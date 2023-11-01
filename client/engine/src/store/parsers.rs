use snafu::{prelude::*, Whatever};
use std::fs;
use std::path::PathBuf;

use super::models;
use super::snapshot::Parser;

pub struct FliptParser {
    namespaces: Vec<String>,
}

// TODO(yquansah): Implement network request here for fetching flag state from upstream
impl Parser for FliptParser {
    fn new(namespaces: Vec<String>) -> Self {
        Self {
            namespaces: namespaces,
        }
    }

    fn parse(&self, _: String) -> Result<models::Document, Whatever> {
        Ok(models::Document {
            version: String::from("1.0"),
            namespace: String::from("default"),
            flags: Vec::new(),
            segments: Vec::new(),
        })
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
        Self {
            namespaces: namespaces,
        }
    }

    fn parse(&self, _: String) -> Result<models::Document, Whatever> {
        let mut d = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
        d.push("src/testdata/state.json");

        let state =
            fs::read_to_string(d.display().to_string()).expect("file should have read correctly");

        let document: models::Document = match serde_json::from_str(&state) {
            Ok(document) => document,
            Err(e) => whatever!("error reading data from server {}", e),
        };

        Ok(document)
    }

    fn get_namespaces(&self) -> Vec<String> {
        self.namespaces.clone()
    }
}
