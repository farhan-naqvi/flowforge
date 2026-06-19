SHELL := /bin/bash
.PHONY: help build test test-unit test-integration lint fmt clean coverage

# Default task
help:
	@echo "FlowForge IR Module - Available targets:"
	@echo "  make build          - Build the module"
	@echo "  make test           - Run all tests"
	@echo "  make test-unit      - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo "  make coverage       - Run tests with coverage"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make fmt-check      - Check formatting"
	@echo "  make clean          - Clean build artifacts"

build:
	go build ./...

test: test-unit test-integration

test-unit:
	go test -v -race ./tests/unit/...

test-integration:
	go test -v ./tests/integration/...

coverage:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	go vet ./...
	which golangci-lint && golangci-lint run ./... || echo "golangci-lint not installed"

fmt:
	go fmt ./...

fmt-check:
	@if [ -n "$$(go fmt ./...)" ]; then echo "Code not formatted"; exit 1; fi
	@echo "Code is properly formatted"

clean:
	rm -f coverage.out coverage.html
	go clean ./...

# Integration: this module builds with go.work
.PHONY: verify-integration
verify-integration:
	@echo "IR module is part of go.work monorepo"
	@echo "Build via root: make -C ../.. build"
