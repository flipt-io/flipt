## Flipt Client Python

This directory contains the Python source code for a Flipt evaluation client using FFI to make calls to a core built in Rust.

### Instructions

To use this client, you can run:

```bash
cargo build --release
```

from the root of the repository and set the `ENGINE_LIB_PATH` environment variable to `{FLIPT_REPO_ROOT}/target/release` which will contain the dynamic linking library necessary for FFI calls with Python.
