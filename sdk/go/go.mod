module go.flipt.io/flipt/sdk/go

go 1.22

require (
	go.flipt.io/flipt/rpc/flipt v1.45.0
	google.golang.org/genproto/googleapis/api v0.0.0-20241007155032-5fefd90f89a9
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241007155032-5fefd90f89a9
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.6.9-0.20230804172637-c7be7c783f49 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	go.flipt.io/flipt/errors v1.45.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto v0.0.0-20240903143218-8af14fe29dc1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	go.flipt.io/flipt/errors => ../../errors/
	go.flipt.io/flipt/rpc/flipt => ../../rpc/flipt/
)
