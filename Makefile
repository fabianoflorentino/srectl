BINARY_NAME=srectl
GO=go
COVERAGE_MIN=80

VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS=-ldflags "-X github.com/fabianoflorentino/srectl/internal/version.Version=$(VERSION) \
	-X github.com/fabianoflorentino/srectl/internal/version.Commit=$(COMMIT) \
	-X github.com/fabianoflorentino/srectl/internal/version.Date=$(BUILD_DATE)"

.PHONY: all build clean test test-short lint lint-fix fmt docker-build install help

.DEFAULT_GOAL := help

## all: Run clean, lint, test and build everything
all: clean lint test build

## build: Build the srectl binary
build:
	$(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/srectl

## clean: Remove build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

## test: Run all tests with the race detector
test:
	$(GO) test -race -coverprofile=coverage.out ./...

## test-short: Run tests without the race detector
test-short:
	$(GO) test -short ./...

## test-integration: Run integration tests against a running cluster
test-integration:
	$(GO) test -tags=integration ./internal/argocd/

## coverage: Report test coverage and fail if below the minimum
coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1
	@total=$$($(GO) tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | tr -d '%'); \
	if [ $$total -lt $(COVERAGE_MIN) ]; then \
		echo "coverage $$total% is below the $(COVERAGE_MIN)% minimum"; exit 1; \
	fi; \
	echo "coverage $$total% >= $(COVERAGE_MIN)% OK"

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## lint-fix: Run golangci-lint and auto-fix
lint-fix:
	golangci-lint run --fix ./...

## fmt: Format the code
fmt:
	$(GO) fmt ./...
	golangci-lint fmt ./...

## docker-build: Build the srectl container image
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

## install: Install srectl to GOBIN
install:
	$(GO) install $(LDFLAGS) ./cmd/srectl

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
