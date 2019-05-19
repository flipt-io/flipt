// +build tools

package main

import (
    _ golang.org/x/tools/cmd/cover
	_ golang.org/x/tools/cmd/goimports
	_ google.golang.org/grpc
	_ github.com/golang/protobuf/protoc-gen-go
	_ github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	_ github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
)