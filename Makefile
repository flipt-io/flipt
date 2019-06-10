PROJECT = flipt

PREFIX ?= $(shell pwd)
SOURCE_FILES ?= ./...

TEST_PATTERN ?= .
TEST_OPTS ?=
TEST_FLAGS ?=

.PHONY: setup
setup: ## Install dev tools
	@echo ">> installing dev tools"
	@if [ ! -f $(GOPATH)/bin/golangci-lint ]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.12.5; \
	fi

.PHONY: test
test: ## Run all the tests
	@echo ">> running tests"
	go test $(TEST_OPTS) -v -covermode=atomic -count=1 -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=30s $(TEST_FLAGS)

.PHONY: cover
cover: test ## Run all the tests and opens the coverage report
	@echo ">> generating test coverage"
	go tool cover -html=coverage.txt

.PHONY: fmt
fmt: ## Run gofmt and goimports on all go files
	@echo ">> running gofmt"
	@find . -name '*.go' -not -wholename './rpc/*' -not -wholename './ui/*' -not -wholename './swagger/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

.PHONY: lint
lint: ## Run all the linters
	@echo ">> running golangci-lint"
	golangci-lint run

.PHONY: clean
clean: ## Remove built binaries
	@echo ">> running go clean"
	go clean -i $(SOURCE_FILES)

.PHONY: proto
proto: ## Build protobufs
	@echo ">> generating protobufs"
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
	@echo ">> generating assets"
	@cd ./ui; yarn install --ignore-engines && yarn run build; cd ..
	go generate ./...

.PHONY: check_assets
check_assets: assets
	@echo ">> checking that assets are up-to-date"
	@if ! (cd ./ui && git diff --exit-code); then \
		echo "run 'make assets' and commit the changes to fix the error."; \
		exit 1; \
	fi
	echo "ok"

.PHONY: build
build: ## Build a local copy
	@echo ">> building a local copy"
	go build -o ./bin/$(PROJECT) ./cmd/$(PROJECT)/.

.PHONY: dev
dev: ## Build and run in development mode
	@echo ">> building and running in development mode"
	go run ./cmd/$(PROJECT)/. --config ./config/local.yml

.PHONY: snapshot
snapshot: ## Build a snapshot version
	@echo ">> building a snapshot version"
	@./build/release/snapshot

.PHONY: release
release: ## Build and publish a release
	@echo ">> building and publishing a release"
	@./build/release/release

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
default: help
