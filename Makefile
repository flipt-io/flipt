PROJECT = flipt

GOTOOLS = \
	golang.org/x/tools/cmd/cover \
	golang.org/x/tools/cmd/goimports \
	google.golang.org/grpc \
	github.com/golang/protobuf/protoc-gen-go \
	github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
	github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
	github.com/golang/dep/cmd/dep \

PREFIX ?= $(shell pwd)
SOURCE_FILES ?= ./...

TEST_PATTERN ?= .
TEST_OPTS ?=

GO111MODULE ?= off

.PHONY: setup
setup: ## Install dev tools
	@if [ ! -f $(GOPATH)/bin/golangci-lint ]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.12.5; \
	fi
	GO111MODULE=$(GO111MODULE) go get $(GOTOOLS)

.PHONY: test
test: ## Run all the tests
	go test $(TEST_OPTS) -v -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=30s

.PHONY: cover
cover: test ## Run all the tests and opens the coverage report
	go tool cover -html=coverage.txt

.PHONY: fmt
fmt: ## Run gofmt and goimports on all go files
	@find . -name '*.go' -not -wholename './proto/*' -not -wholename './vendor/*' -not -wholename './ui/*' -not -wholename './swagger/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

.PHONY: lint
lint: ## Run all the linters
	golangci-lint run

.PHONY: clean
clean: ## Remove built binaries
	go clean -i $(SOURCE_FILES)
	rm -rf ./bin/* ./dist/*

.PHONY: proto
proto: ## Build protobufs
	protoc -I/usr/local/include -I. \
		-I $(GOPATH)/src \
		-I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway \
		-I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I proto \
		--go_out=plugins=grpc:./proto \
		--grpc-gateway_out=logtostderr=true,grpc_api_configuration=./proto/flipt.yaml:./proto \
		--swagger_out=logtostderr=true,grpc_api_configuration=./proto/flipt.yaml:./swagger/api \
		$(PROJECT).proto

.PHONY: ui
ui: ## Builds the ui
	@cd ./ui; yarn install && yarn run build

.PHONY: generate
generate: ## Run go generate
	GO111MODULE=$(GO111MODULE) go generate ./...

.PHONY: build
build: ## Build a local copy
	GO111MODULE=$(GO111MODULE) go build -o ./bin/$(PROJECT) ./cmd/$(PROJECT)/main.go

.PHONY: dev
dev: ## Build and run in development mode
	GO111MODULE=$(GO111MODULE) go run -tags=dev ./cmd/$(PROJECT)/main.go --config ./config/local.yml

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
default: help
