module go.flipt.io/flipt/sdk/go/v2

go 1.23.0

toolchain go1.24.2

require (
	go.flipt.io/flipt/rpc/v2/environments v0.0.0-00010101000000-000000000000
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250428153025-10db94c68c34
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic v0.7.0 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	go.flipt.io/flipt/errors v1.45.0 // indirect
	go.flipt.io/flipt/rpc/flipt v1.54.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250505200425-f936aa4a68b2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	go.flipt.io/flipt/errors => ../../../errors/
	go.flipt.io/flipt/rpc/flipt => ../../../rpc/flipt/
	go.flipt.io/flipt/rpc/v2/environments => ../../../rpc/v2/environments/
)
