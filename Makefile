PROJECT = flipt

SOURCE_FILES ?= ./...

BENCH_PATTERN ?= .
TEST_PATTERN ?= .
TEST_OPTS ?=
TEST_FLAGS ?= -v

GOBIN = _tools/bin
export PATH := $(GOBIN):$(PATH)

UI_PATH = ui
UI_SOURCE_FILES = $(wildcard $(UI_PATH)/static/* $(UI_PATH)/src/**/* $(UI_PATH)/src/**/**/* $(UI_PATH)/index.html)
UI_NODE_MODULES_PATH = $(UI_PATH)/node_modules
UI_OUTPUT_PATH = $(UI_PATH)/dist

$(UI_NODE_MODULES_PATH): $(UI_PATH)/package.json $(UI_PATH)/yarn.lock
	@cd $(UI_PATH) && yarn --frozen-lockfile

$(UI_OUTPUT_PATH): $(UI_NODE_MODULES_PATH) $(UI_SOURCE_FILES)
	@cd $(UI_PATH) && yarn build

.PHONY: bootstrap
bootstrap: ## Install dev tools
	@echo ">> installing dev tools"
	@./script/bootstrap

.PHONY: bench
bench: ## Run all the benchmarks
	@echo ">> running benchmarks"
	go test -bench=$(BENCH_PATTERN) $(SOURCE_FILES) -run=XXX $(TEST_FLAGS)

.PHONY: test
test: ## Run all the tests
	@echo ">> running tests"
	go test $(TEST_OPTS) -covermode=atomic -count=1 -coverprofile=coverage.txt $(SOURCE_FILES) -run=$(TEST_PATTERN) -timeout=30s $(TEST_FLAGS)

.PHONY: cover
cover: test ## Run all the tests and opens the coverage report
	@echo ">> generating test coverage"
	go tool cover -html=coverage.txt

.PHONY: fmt
fmt: ## Run gofmt and goimports on all go files
	@echo ">> running gofmt"
	@find . -name '*.go' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

.PHONY: lint
lint: ## Run all the linters
	@echo ">> running golangci-lint"
	golangci-lint run

.PHONY: clean
clean: ## Cleanup generated files
	@echo ">> cleaning up"
	go clean -i $(SOURCE_FILES)
	rm -rf dist/*
	go mod tidy

.PHONY: proto
proto: ## Build protobufs
	@echo ">> generating protobufs"
	buf generate

.PHONY: assets
assets: $(UI_OUTPUT_PATH) ## Build the ui

commit-hash=$(shell set -e && git rev-parse --verify HEAD || "")

.PHONY: build
build: clean assets ## Build a local copy
	@echo ">> building a local copy"
	go build -tags assets -ldflags "-X main.commit=$(commit-hash)" -o ./bin/$(PROJECT) ./cmd/$(PROJECT)/.

.PHONY: server
server: clean  ## Build and run in server mode
	@echo ">> building and running in server mode"
	@echo "  ⚠️ ui must be run in another process ⚠️"
	go run ./cmd/$(PROJECT)/. --config ./config/local.yml --force-migrate

.PHONY: snapshot
snapshot: clean assets ## Build a snapshot version
	@echo ">> building a snapshot version"
	@./script/build snapshot

.PHONY: release
release: clean assets ## Build and publish a release
	@echo ">> building and publishing a release"
	@./script/build release

.PHONY: clients
clients: ## Generate GRPC clients
	@echo ">> generating GRPC clients"
	@buf generate --template=buf.public.gen.yaml

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build
default: build
