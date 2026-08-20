module go.flipt.io/flipt/rpc/flipt

go 1.25.0

toolchain go1.27.0

require (
	github.com/google/gnostic-models v0.7.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0
	github.com/stretchr/testify v1.12.1
	go.flipt.io/flipt/errors v1.45.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260819154853-08b0e4226688
	google.golang.org/grpc v1.83.0
	google.golang.org/protobuf v1.36.12
)

require (
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260818201246-1b0934165a6f // indirect
)

replace go.flipt.io/flipt/errors => ../../errors/
