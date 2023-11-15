# Flipt Client Engine

![Status: Experimental](https://img.shields.io/badge/status-experimental-yellow)

This is the client engine for Flipt. It is responsible for evaluating context provided by the native language client SDKs and returning the results of the evaluation.

## Architecture

The client engine is a Rust library responsible for evaluating context and returning the results of the evaluation. The client engine polls for evaluation state from the Flipt server and uses this state to determine the results of the evaluation. The client engine is designed to be embedded in the native language client SDKs. The native language client SDKs will send context to the client engine via [FFI](https://en.wikipedia.org/wiki/Foreign_function_interface) and receive the results of the evaluation from engine.

This design allows for the client evaluation logic to be written once in a memory safe language and embedded in the native language client SDKs. This design also allows for the client engine to be updated independently of the native language client SDKs.

TODO: Diagram

## Development

TODO: Add more details

### Prerequisites

- [Rust](https://www.rust-lang.org/tools/install)
- [cbindgen](https://github.com/mozilla/cbindgen)

### Building the Library

Development:

```bash
cargo build
```

Specify a relative target directory:

```bash
cargo build --target-dir ./target
```

Release:

```bash
cargo build --release
```

Specify a relative target directory:

```bash
cargo build --release --target-dir ./target
```

### Test the Library

```bash
cargo test
```

### Generate the FFI Header

Requires [cbindgen](https://github.com/mozilla/cbindgen)

```bash
cbindgen --config cbindgen.toml --crate flipt-client-engine --output flipt_engine.h
```
