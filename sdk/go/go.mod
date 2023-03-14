module go.flipt.io/flipt/sdk

go 1.18

require (
	go.flipt.io/flipt/rpc/flipt v0.0.0-00010101000000-000000000000
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.2 // indirect
	go.flipt.io/flipt/errors v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/grpc v1.53.0 // indirect
)

replace (
	go.flipt.io/flipt/errors => ../../errors/
	go.flipt.io/flipt/rpc/flipt => ../../rpc/flipt/
)
