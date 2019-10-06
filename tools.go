// +build tools

package tools

import (
	_ "github.com/gobuffalo/packr/packr"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"
	_ "golang.org/x/tools/cmd/cover"
	_ "golang.org/x/tools/cmd/goimports"
	_ "google.golang.org/grpc"
)
