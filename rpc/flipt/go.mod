module go.flipt.io/flipt/rpc/flipt

go 1.21

require (
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1
	github.com/stretchr/testify v1.9.0
	go.flipt.io/flipt/errors v1.19.2
	go.uber.org/zap v1.27.0
	google.golang.org/genproto/googleapis/api v0.0.0-20240325203815-454cdb8f5daa
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240311173647-c811ad7063a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240325203815-454cdb8f5daa // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace go.flipt.io/flipt/errors => ../../errors/
