PROJECT = flipt

SOURCE_FILES ?= ./...

TEST_PATTERN ?= .
TEST_OPTS ?=
TEST_FLAGS ?=

TOOLS = \
	"github.com/gobuffalo/packr/packr" \
	"github.com/golang/protobuf/protoc-gen-go" \
	"github.com/golangci/golangci-lint/cmd/golangci-lint" \
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway" \
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger" \
	"golang.org/x/tools/cmd/cover" \
	"golang.org/x/tools/cmd/goimports" \
	"google.golang.org/grpc" \

.PHONY: setup
setup: ## Install dev tools
	@echo ">> installing dev tools"
	go install -v $(TOOLS)

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
	packr clean

.PHONY: proto
proto: ## Build protobufs
	@echo ">> generating protobufs"
	protoc -I/usr/local/include -I. \
		-I rpc \
		--go_out=plugins=grpc:./rpc \
		--grpc-gateway_out=logtostderr=true,grpc_api_configuration=./rpc/flipt.yaml:./rpc \
		--swagger_out=logtostderr=true,grpc_api_configuration=./rpc/flipt.yaml:./swagger/api \
		$(PROJECT).proto

.PHONY: assets
assets: ## Build the ui and run go generate
	@echo ">> generating assets"
	@cd ./ui; yarn install --ignore-platform --ignore-engines && yarn run build; cd ..
	packr -v

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
	@./script/build/snapshot

.PHONY: release
release: ## Build and publish a release
	@echo ">> building and publishing a release"
	@./script/build/release

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
default: help
