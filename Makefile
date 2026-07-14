# FlowForge — real targets only. Everything here is verified to run.
# The product is the Python package `flowforge/`; `ir/` is the Go schema-of-record.

.PHONY: help setup test test-go smoke demo serve lint clean

PY ?= python3

help:
	@echo "FlowForge — available targets:"
	@echo "  make setup     Install the package (editable) with dev extras"
	@echo "  make test      Run the Python test suite (pytest)"
	@echo "  make test-go   Build & test the Go IR schema-of-record"
	@echo "  make smoke     Validate + compile the example pipelines"
	@echo "  make demo      Run the guided end-to-end demo (docs/DEMO.md)"
	@echo "  make serve     Start the local workbench UI on 127.0.0.1:8080"
	@echo "  make lint      go vet ./ir/... (Python is dependency-light)"
	@echo "  make clean     Remove caches and local run stores"

setup:
	$(PY) -m pip install -e ".[dev]"

test:
	$(PY) -m pytest -q

test-go:
	cd ir && go build ./... && go vet ./... && go test ./...

smoke:
	@for f in examples/simple_etl.py examples/fan_out_fan_in.py; do \
		echo "== $$f =="; \
		$(PY) -m flowforge.cli validate $$f; \
		$(PY) -m flowforge.cli compile $$f argo -o /tmp/$$(basename $$f).argo.yaml; \
		$(PY) -m flowforge.cli compile $$f airflow -o /tmp/$$(basename $$f).airflow.py; \
	done
	@echo "checking broken pipeline is rejected:"
	@$(PY) -m flowforge.cli validate examples/broken_cycle.py; test $$? -ne 0 && echo "  (correctly rejected)"

demo:
	$(PY) scripts/demo.py

serve:
	$(PY) -m flowforge.cli serve --workspace examples

lint:
	cd ir && go vet ./...

clean:
	find . -name '__pycache__' -type d -prune -exec rm -rf {} + 2>/dev/null || true
	find . -name '*.pyc' -delete 2>/dev/null || true
	rm -rf .pytest_cache *.egg-info build dist
	rm -f *.db examples/*.db
