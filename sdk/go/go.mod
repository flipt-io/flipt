module go.flipt.io/flipt/sdk/go

go 1.24.0

toolchain go1.25.7

require (
	go.flipt.io/flipt/rpc/flipt v1.54.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260203192932-546029d2fa20
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260128011058-8636f8732409
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.7 // indirect
	go.flipt.io/flipt/errors v1.45.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.33.0 // indirect
)

replace (
	go.flipt.io/flipt/errors => ../../errors/
	go.flipt.io/flipt/rpc/flipt => ../../rpc/flipt/
)
