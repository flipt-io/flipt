module go.flipt.io/flipt/rpc/flipt

go 1.25.0

toolchain go1.27.0

require (
	github.com/google/gnostic-models v0.7.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0
	github.com/stretchr/testify v1.12.1
	go.flipt.io/flipt/errors v1.45.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260831171406-18b4a7587f8a
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12
)

require (
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260825221802-da73d73af1c5 // indirect
)

replace go.flipt.io/flipt/errors => ../../errors/
