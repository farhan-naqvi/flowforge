# FlowForge Monorepo Makefile
# Common tasks for all modules

.PHONY: help setup-dev build test test-unit test-integration test-e2e test-coverage lint fmt lint-fix proto-gen dev-up dev-down dev-logs build-images

SHELL := /bin/bash
GO_MODULES := ir compiler runtime executors storage lineage api observability
PYTHON_MODULES := sdk

help:
	@echo "FlowForge Monorepo - Available Commands"
	@echo ""
	@echo "Setup & Development:"
	@echo "  make setup-dev              Setup local development environment"
	@echo "  make dev-up                 Start docker-compose (services)"
	@echo "  make dev-down               Stop services"
	@echo "  make dev-logs               View service logs"
	@echo ""
	@echo "Build:"
	@echo "  make build                  Build all modules (Go + Python)"
	@echo "  make build-images           Build Docker images"
	@echo ""
	@echo "Test:"
	@echo "  make test                   Run all tests (unit + integration + e2e)"
	@echo "  make test-unit              Run unit tests only"
	@echo "  make test-integration       Run integration tests"
	@echo "  make test-e2e               Run e2e tests"
	@echo "  make test-coverage          Generate coverage reports"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint                   Lint all code"
	@echo "  make fmt                    Format all code"
	@echo "  make lint-fix               Auto-fix linting issues"
	@echo ""
	@echo "Code Generation:"
	@echo "  make proto-gen              Generate protobuf code"
	@echo "  make api-docs               Generate OpenAPI docs"
	@echo ""

# ============================================================================
# Setup & Development
# ============================================================================

setup-dev:
	@echo "🔧 Setting up development environment..."
	@command -v go >/dev/null 2>&1 || (echo "❌ Go is required. Install from https://golang.org/doc/install" && exit 1)
	@command -v python3 >/dev/null 2>&1 || (echo "❌ Python 3 is required. Install from https://www.python.org/downloads/" && exit 1)
	@command -v docker >/dev/null 2>&1 || (echo "❌ Docker is required. Install from https://docs.docker.com/get-docker/" && exit 1)
	@command -v docker-compose >/dev/null 2>&1 || (echo "❌ Docker Compose is required. Install from https://docs.docker.com/compose/install/" && exit 1)
	@echo "✓ All dependencies installed"
	@echo "🐍 Installing Python dependencies..."
	cd sdk && pip install -e ".[dev]" && cd ..
	@echo "✓ Development environment ready"
	@echo ""
	@echo "Next steps:"
	@echo "  make dev-up    # Start services"
	@echo "  make test      # Run tests"

dev-up:
	@echo "🚀 Starting development services..."
	docker-compose up -d
	@echo "✓ Services started"
	@echo ""
	@echo "Services:"
	@echo "  API:       http://localhost:8080"
	@echo "  UI:        http://localhost:3000"
	@echo "  Postgres:  localhost:5432"
	@echo "  Redis:     localhost:6379"
	@echo "  Argo:      http://localhost:2746"
	@echo ""
	@docker-compose logs -f

dev-down:
	@echo "🛑 Stopping development services..."
	docker-compose down
	@echo "✓ Services stopped"

dev-logs:
	docker-compose logs -f

# ============================================================================
# Build
# ============================================================================

build: build-go build-python
	@echo "✓ All modules built"

build-go:
	@echo "🔨 Building Go modules..."
	@for module in $(GO_MODULES); do \
		echo "  Building $$module..."; \
		cd $$module && go build ./... && cd .. || exit 1; \
	done
	@echo "✓ Go modules built"

build-python:
	@echo "🐍 Building Python SDK..."
	cd sdk && python setup.py build && cd ..
	@echo "✓ Python SDK built"

build-images:
	@echo "🐳 Building Docker images..."
	docker-compose build
	@echo "✓ Docker images built"

# ============================================================================
# Test
# ============================================================================

test: test-unit test-integration test-e2e
	@echo "✓ All tests passed"

test-unit:
	@echo "🧪 Running unit tests..."
	@for module in $(GO_MODULES); do \
		echo "  Testing $$module (unit)..."; \
		cd $$module && go test -short -v -cover ./... && cd .. || exit 1; \
	done
	cd sdk && python -m pytest tests/unit -v && cd ..
	@echo "✓ Unit tests passed"

test-integration:
	@echo "🔗 Running integration tests..."
	@for module in $(GO_MODULES); do \
		echo "  Testing $$module (integration)..."; \
		cd $$module && go test -run Integration -v ./... && cd .. || exit 1; \
	done
	cd sdk && python -m pytest tests/integration -v && cd ..
	@echo "✓ Integration tests passed"

test-e2e:
	@echo "🌐 Running e2e tests..."
	cd tests && python -m pytest e2e -v && cd ..
	@echo "✓ E2E tests passed"

test-coverage:
	@echo "📊 Generating coverage reports..."
	@mkdir -p coverage
	@for module in $(GO_MODULES); do \
		echo "  Coverage for $$module..."; \
		cd $$module && go test -coverprofile=../coverage/$$module.out ./... && cd .. || exit 1; \
	done
	cd sdk && python -m pytest --cov=flowforge --cov-report=html tests/ && cd ..
	@echo "✓ Coverage reports generated"
	@echo "  Go:     coverage/ directory"
	@echo "  Python: sdk/htmlcov/index.html"

# ============================================================================
# Code Quality
# ============================================================================

lint: lint-go lint-python
	@echo "✓ All linting passed"

lint-go:
	@echo "🔍 Linting Go code..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@for module in $(GO_MODULES); do \
		echo "  Linting $$module..."; \
		cd $$module && golangci-lint run ./... && cd .. || exit 1; \
	done
	@echo "✓ Go linting passed"

lint-python:
	@echo "🔍 Linting Python code..."
	cd sdk && python -m pylint flowforge && cd ..
	cd ui && npm run lint && cd ..
	@echo "✓ Python/TS linting passed"

fmt:
	@echo "🎨 Formatting all code..."
	@for module in $(GO_MODULES); do \
		echo "  Formatting $$module..."; \
		cd $$module && go fmt ./... && cd .. || exit 1; \
	done
	cd sdk && black flowforge && cd ..
	cd ui && npm run format && cd ..
	@echo "✓ All code formatted"

lint-fix:
	@echo "🔧 Auto-fixing linting issues..."
	@for module in $(GO_MODULES); do \
		echo "  Fixing $$module..."; \
		cd $$module && golangci-lint run --fix ./... && cd .. || exit 1; \
	done
	cd sdk && black flowforge --quiet && cd ..
	cd ui && npm run format && cd ..
	@echo "✓ Linting issues fixed"

# ============================================================================
# Code Generation
# ============================================================================

proto-gen:
	@echo "🔧 Generating protobuf code..."
	@command -v protoc >/dev/null 2>&1 || (echo "Installing protoc..." && brew install protobuf)
	cd api && \
		protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. proto/src/flowforge/**/*.proto && \
		echo "✓ Protobuf code generated"

api-docs:
	@echo "📄 Generating OpenAPI docs..."
	cd api && protoc --openapiv2_out=. proto/src/flowforge/**/*.proto && cd ..
	@echo "✓ OpenAPI docs generated"

# ============================================================================
# Directory Structure Setup (for new projects)
# ============================================================================

dirs:
	@echo "📁 Creating directory structure..."
	@for module in $(GO_MODULES); do \
		mkdir -p $$module/{internal,pkg,tests/{unit,integration,fixtures},cmd} && \
		touch $$module/README.md $$module/Makefile $$module/go.mod $$module/go.sum; \
	done
	@mkdir -p sdk/{flowforge/{cli/commands,validators,utils},tests/{unit,integration,fixtures}}
	@mkdir -p ui/{src/{components,pages,services,hooks,store,types,utils,styles},tests/{unit,integration,e2e,fixtures},public}
	@mkdir -p deployment/{terraform/{modules,environments},helm/flowforge/{templates,dependencies},docker,scripts}
	@mkdir -p examples/{basic,data_quality,ml_workflow,multi_executor}/transformations
	@mkdir -p tests/{e2e,fixtures/{pipelines,expected_outputs}}
	@mkdir -p scripts docs/.github/{workflows,ISSUE_TEMPLATE}
	@echo "✓ Directory structure created"

# ============================================================================
# Utility
# ============================================================================

version:
	@cat VERSION

clean:
	@echo "🗑️  Cleaning build artifacts..."
	@for module in $(GO_MODULES); do \
		cd $$module && go clean && cd .. || exit 1; \
	done
	rm -rf coverage/ dist/ build/
	@echo "✓ Cleaned"

.DEFAULT_GOAL := help
