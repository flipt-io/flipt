PROJECT = flipt

PREFIX ?= $(shell pwd)
SOURCE_FILES ?= ./...

TEST_PATTERN ?= .
TEST_OPTS ?=
TEST_FLAGS ?=

.PHONY: setup
setup: ## Install dev tools
	@if [ ! -f $(GOPATH)/bin/golangci-lint ]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.12.5; \
	fi

.PHONY: test
test: ## Run all the tests
	go test $(TEST_OPTS) -v -covermode=atomic -count=1 -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=30s $(TEST_FLAGS)

.PHONY: cover
cover: test ## Run all the tests and opens the coverage report
	go tool cover -html=coverage.txt

.PHONY: fmt
fmt: ## Run gofmt and goimports on all go files
	@find . -name '*.go' -not -wholename './rpc/*' -not -wholename './ui/*' -not -wholename './swagger/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

.PHONY: lint
lint: ## Run all the linters
	golangci-lint run

.PHONY: clean
clean: ## Remove built binaries
	go clean -i $(SOURCE_FILES)

.PHONY: proto
proto: ## Build protobufs
	protoc -I/usr/local/include -I. \
		-I $(GOPATH)/src \
		-I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway \
		-I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I rpc \
		--go_out=plugins=grpc:./rpc \
		--grpc-gateway_out=logtostderr=true,grpc_api_configuration=./rpc/flipt.yaml:./rpc \
		--swagger_out=logtostderr=true,grpc_api_configuration=./rpc/flipt.yaml:./swagger/api \
		$(PROJECT).proto

.PHONY: assets
assets: ## Build the ui and run go generate
	@cd ./ui; yarn install && yarn run build; cd ..
	go generate ./...

.PHONY: build
build: ## Build a local copy
	go build -o ./bin/$(PROJECT) ./cmd/$(PROJECT)/.

.PHONY: dev
dev: ## Build and run in development mode
	go run ./cmd/$(PROJECT)/. --config ./config/local.yml

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
default: help
