.PHONY: help install dev test test-unit test-integration coverage lint format clean

help:
	@echo "FlowForge SDK - Available commands:"
	@echo "  make install        Install package"
	@echo "  make dev            Install with dev dependencies"
	@echo "  make test           Run all tests"
	@echo "  make test-unit      Run unit tests only"
	@echo "  make test-int       Run integration tests only"
	@echo "  make coverage       Run tests with coverage report"
	@echo "  make lint           Run linter (pylint)"
	@echo "  make format         Auto-format code (black + isort)"
	@echo "  make clean          Remove build artifacts"

install:
	pip install -e .

dev:
	pip install -e ".[dev,aws,spark]"

test:
	pytest tests/ -v

test-unit:
	pytest tests/unit/ -v

test-int:
	pytest tests/integration/ -v

coverage:
	pytest tests/ --cov=flowforge --cov-report=term-missing --cov-report=html
	@echo "Coverage report: htmlcov/index.html"

lint:
	pylint flowforge/

format:
	black flowforge/ tests/
	isort flowforge/ tests/

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name "*.egg-info" -exec rm -rf {} + 2>/dev/null || true
	rm -rf build/ dist/ .coverage htmlcov/ .pytest_cache/ .mypy_cache/
