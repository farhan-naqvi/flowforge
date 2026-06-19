# flowforge CLI

Standalone CLI for FlowForge pipeline parsing, validation, and code generation.

## Install

```bash
cd flowforge-cli
pip install .
```

## Usage

```bash
flowforge generate argo examples/simple_etl.py
flowforge generate airflow examples/simple_etl.py
flowforge validate examples/simple_etl.py
flowforge preview examples/simple_etl.py
```

## Commands

- `flowforge generate argo <pipeline.py> [--output path]`
- `flowforge generate airflow <pipeline.py> [--output path]`
- `flowforge validate <pipeline.py>`
- `flowforge preview <pipeline.py>`

## Supported syntax

- `@pipeline(...)`
- `@task(...)`
- `<pipeline>.add_task("id", function_name)`
- `<pipeline>.add_edge(src_fn, "src_port", dst_fn, "dst_port")`
