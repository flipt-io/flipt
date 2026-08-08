module go.flipt.io/flipt/sdk/go

go 1.25.0

toolchain go1.26.5

require (
	go.flipt.io/flipt/rpc/flipt v1.54.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260807164820-c8921c73eeea
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260803160001-6ac0973c030d
	google.golang.org/grpc v1.83.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 // indirect
	go.flipt.io/flipt/errors v1.45.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace (
	go.flipt.io/flipt/errors => ../../errors/
	go.flipt.io/flipt/rpc/flipt => ../../rpc/flipt/
)
