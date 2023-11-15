## Flipt Client Go

The `flipt-client-go` directory contains the Golang source code for a Flipt evaluation client using FFI to make calls to a core built in Rust.

### Instructions

To use this client, you can run the following command from the root of the repository:

```bash
cargo build --release
```

This should generate a `target/` directory in the root of this repository, which contains the dynamic linking library built for your platform. This dynamic library will contain the functionality necessary for the Golang client to make FFI calls.

You can import the module that contains the evaluation client: `go.flipt.io/flipt/flipt-client-go` and build your Go project with the `CGO_LDFLAGS` environment variable set:

```bash
CGO_LDFLAGS="-L/path/to/lib -lfliptengine"
```

The `path/to/lib` will be the path to the dynamic library which will have the following paths depending on your platform.

- **Linux**: `{FLIPT_REPO_ROOT}/target/release/libfliptengine.so`
- **Windows**: `{FLIPT_REPO_ROOT}/target/release/libfliptengine.dll`
- **MacOS**: `{FLIPT_REPO_ROOT}/target/release/libfliptengine.dylib`

You can then use the client like so:

```golang
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	evaluation "go.flipt.io/flipt/flipt-client-go"
)

func main() {
	evaluationClient := evaluation.NewClient("default")

	evalCtx := map[string]string{
		"fizz": "buzz",
	}

	evalCtxBytes, err := json.Marshal(evalCtx)
	if err != nil {
		log.Fatal(err)
	}

	variantEvaluationResponse, err := evaluationClient.Variant(context.Background(), &evaluation.EvaluationRequest{
		NamespaceKey: "default",
		FlagKey:      "flag1",
		EntityId:     "someentity",
		Context:      string(evalCtxBytes),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(variantEvaluationResponse)
}
```
