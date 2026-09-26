module go.flipt.io/flipt/rpc/flipt

go 1.26.0

toolchain go1.27.1

require (
	github.com/google/gnostic-models v0.7.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.31.0
	github.com/stretchr/testify v1.12.1
	go.flipt.io/flipt/errors v1.45.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260921155816-b14227669459
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
)

require (
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260918162117-cecb64721679 // indirect
)

replace go.flipt.io/flipt/errors => ../../errors/
