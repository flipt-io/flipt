module go.flipt.io/flipt/sdk/go

go 1.24.0

toolchain go1.25.2

require (
	go.flipt.io/flipt/rpc/flipt v1.54.0
	google.golang.org/genproto/googleapis/api v0.0.0-20251007200510-49b9836ed3ff
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251002232023-7c0ddcbb5797
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	go.flipt.io/flipt/errors v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

replace (
	go.flipt.io/flipt/errors => ../../errors/
	go.flipt.io/flipt/rpc/flipt => ../../rpc/flipt/
)
