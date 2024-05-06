module go.flipt.io/flipt/sdk/go

go 1.22

require (
	go.flipt.io/flipt/rpc/flipt v1.38.0
	google.golang.org/genproto/googleapis/api v0.0.0-20240325203815-454cdb8f5daa
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240325203815-454cdb8f5daa
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	go.flipt.io/flipt/errors v1.19.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240311173647-c811ad7063a7 // indirect
)

replace (
	go.flipt.io/flipt/errors => ../../errors/
	go.flipt.io/flipt/rpc/flipt => ../../rpc/flipt/
)
