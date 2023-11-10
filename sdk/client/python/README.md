## Flipt Client Python

This directory contains the Python source code for a Flipt evaluation client using FFI to make calls to a core built in Rust.

### Instructions

To use this client, you can run the following command from the root of the repository:

```bash
cargo build
```

This should generate a `target/` directory in the root of this repository, which contains the dynamic linking library built for your platform. This dynamic library will contain the functionality necessary for the Python client to make FFI calls. You'll need to set the `ENGINE_LIB_PATH` environment variable depending on your platform:

- **Linux**: `{FLIPT_REPO_ROOT}/target/debug/libengine.so`
- **Windows**: `{FLIPT_REPO_ROOT}/target/debug/libengine.dll`
- **MacOS**: `{FLIPT_REPO_ROOT}/target/debug/libengine.dylib`

In your Python code you can import this client and use it as so:

```python
from flipt_client_python import FliptEvaluationClient

flipt_evaluation_client = FliptEvaluationClient(namespaces=["default", "another-namespace"])

variant_response = flipt_evaluation_client.variant(namespace_key="default", flag_key="flag1", entity_id="entity", context={"this": "context"})
```

