module go.flipt.io/flipt/rpc/flipt

go 1.24.0

toolchain go1.25.3

require (
	github.com/google/gnostic-models v0.7.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/stretchr/testify v1.11.1
	go.flipt.io/flipt/errors v1.45.0
	google.golang.org/genproto/googleapis/api v0.0.0-20251029180050-ab9386a59fda
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251014184007-4626949a642f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace go.flipt.io/flipt/errors => ../../errors/
