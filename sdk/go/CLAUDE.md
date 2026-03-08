# AGENTS — Go SDK

## Overview

The Go SDK provides a typed client for the Flipt API with two transport implementations (gRPC and HTTP). There are two separate SDK modules:

- **v1 SDK** (`sdk/go/`) — Wraps the v1 API (`rpc/flipt/`): flags, segments, evaluation, authentication
- **v2 SDK** (`sdk/go/v2/`) — Wraps the v2 API (`rpc/v2/environments/`): environments, namespaces, resources

Each module is a separate Go module with its own `go.mod`.

## Important: No Code Generation

The `.sdk.gen.go` files in this directory were **originally generated** by `protoc-gen-go-flipt-sdk` but are **now maintained by hand**. The `// Code generated` header is historical — do not attempt to regenerate these files. When new RPCs are added to the proto definitions, the corresponding SDK methods must be added manually.

## Architecture

Each SDK module has three layers:

```
sdk/go/[v2/]
├── *.sdk.gen.go          # SDK wrapper — authenticates and delegates to transport
├── sdk.gen.go            # Transport interface + SDK constructor
├── grpc/
│   └── grpc.sdk.gen.go   # gRPC transport — wraps the generated gRPC client
└── http/
    ├── http.sdk.gen.go    # HTTP transport — shared utilities (checkResponse, etc.)
    └── *.sdk.gen.go       # HTTP transport — per-service HTTP client methods
```

### Layer 1: SDK Wrapper (`environments.sdk.gen.go`, `flipt.sdk.gen.go`, etc.)

Thin wrapper that handles authentication and delegates to the transport:

```go
func (x *Environments) CopyResource(ctx context.Context, v *environments.CopyResourceRequest) (*environments.CopyResourceResponse, error) {
	ctx, err := authenticate(ctx, x.authenticationProvider)
	if err != nil {
		return nil, err
	}
	return x.transport.CopyResource(ctx, v)
}
```

Every method follows this exact pattern: authenticate, then delegate.

### Layer 2: gRPC Transport (`grpc/grpc.sdk.gen.go`)

Minimal — just creates the gRPC client from a connection. No per-method code needed because the generated gRPC client already implements the full interface:

```go
func (t Transport) EnvironmentsClient() environments.EnvironmentsServiceClient {
	return environments.NewEnvironmentsServiceClient(t.cc)
}
```

**You do not need to modify gRPC transport files when adding new RPCs.** The generated gRPC client interface (`EnvironmentsServiceClient`) is updated automatically by `buf generate`.

### Layer 3: HTTP Transport (`http/environments.sdk.gen.go`, etc.)

Each RPC needs an explicit HTTP method mapping. Follow the existing pattern:

- **GET** endpoints: Build URL with path parameters, no request body
- **POST/PUT** endpoints: Marshal request to JSON, send as body
- **DELETE** endpoints: Set query parameters (e.g., `revision`), no body

```go
func (x *EnvironmentsServiceClient) CopyResource(ctx context.Context, v *environments.CopyResourceRequest, _ ...grpc.CallOption) (*environments.CopyResourceResponse, error) {
	var body io.Reader
	var values url.Values
	reqData, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	body = bytes.NewReader(reqData)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.addr+fmt.Sprintf("/api/v2/environments/%v/namespaces/%v/resources/copy", v.EnvironmentKey, v.NamespaceKey), body)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()
	resp, err := x.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var output environments.CopyResourceResponse
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, respData); err != nil {
		return nil, err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(respData, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
```

The URL path must match the `google.api.http` annotation in the proto file and the HTTP method must match (`post` = `http.MethodPost`, `get` = `http.MethodGet`, etc.).

## Adding a New RPC

When a new RPC is added to a proto service:

### Step 1: Regenerate proto/gRPC code

```bash
mise run proto
```

This updates the generated gRPC client interface — no SDK changes needed for gRPC transport.

### Step 2: Add SDK wrapper method

**v2 example:** Edit `sdk/go/v2/environments.sdk.gen.go` — add a method to the `Environments` struct following the authenticate-then-delegate pattern.

**v1 example:** Edit the appropriate file in `sdk/go/` (e.g., `flipt.sdk.gen.go` for flag operations).

### Step 3: Add HTTP transport method

**v2:** Edit `sdk/go/v2/http/environments.sdk.gen.go`
**v1:** Edit the appropriate file in `sdk/go/http/`

Build the URL from the proto's `google.api.http` annotation. Use the correct HTTP method.

### Step 4: Verify

```bash
# v2
cd sdk/go/v2 && go build ./...

# v1
cd sdk/go && go build ./...
```

## File Reference

### v1 SDK (`sdk/go/`)

| File | Purpose |
|------|---------|
| `sdk.gen.go` | `Transport` interface, `SDK` struct, `New()` constructor |
| `flipt.sdk.gen.go` | `Flipt` struct — flag/segment CRUD methods |
| `evaluation.sdk.gen.go` | `Evaluation` struct — flag evaluation methods |
| `auth.sdk.gen.go` | `Auth` struct — authentication methods |
| `grpc/grpc.sdk.gen.go` | gRPC transport implementation |
| `http/http.sdk.gen.go` | HTTP transport utilities |
| `http/flipt.sdk.gen.go` | HTTP transport for flag/segment operations |
| `http/evaluation.sdk.gen.go` | HTTP transport for evaluation |
| `http/auth.sdk.gen.go` | HTTP transport for authentication |

### v2 SDK (`sdk/go/v2/`)

| File | Purpose |
|------|---------|
| `sdk.gen.go` | `Transport` interface, `SDK` struct, `New()` constructor |
| `environments.sdk.gen.go` | `Environments` struct — all environment/namespace/resource methods |
| `grpc/grpc.sdk.gen.go` | gRPC transport implementation |
| `http/http.sdk.gen.go` | HTTP transport utilities |
| `http/environments.sdk.gen.go` | HTTP transport for all environment operations |

### Adding a new v2 service

If a completely new v2 service is added (not just new RPCs on `EnvironmentsService`):

1. Add a new `*Client()` method to the `Transport` interface in `sdk/go/v2/sdk.gen.go`
2. Create a new `<service>.sdk.gen.go` with the wrapper struct and methods
3. Add a factory method on `SDK` (e.g., `func (s SDK) NewService() *NewService`)
4. Add the client creation to `grpc/grpc.sdk.gen.go`
5. Create `http/<service>.sdk.gen.go` with HTTP transport methods
6. Add client creation to `http/http.sdk.gen.go`
