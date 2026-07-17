# FlowForge VS Code Extension

FlowForge converts annotated Python pipeline files into:
- Argo Workflows YAML (`.argo.yaml`)
- Apache Airflow DAG Python (`.airflow.py`)

## Supported decorators and calls

```python
@pipeline(name="etl", version="1.0.0", owner="data_team")
def my_pipeline():
    pass

@task(image="python:3.11", timeout="3600s", retries=3)
def extract() -> list:
    return []

@task(image="python:3.11")
def transform(data: list) -> list:
    return []

my_pipeline.add_task("extract", extract)
my_pipeline.add_task("transform", transform)
my_pipeline.add_edge(extract, "result", transform, "data")
```

## Commands

- `Generate Argo Workflow`
- `Generate Airflow DAG`
- `Preview FlowForge DAG`

Commands are available from Python file editor/explorer context menus.

## Build

```bash
cd flowforge-vscode
npm install
npm run compile
```

Press `F5` in VS Code to run the extension in an Extension Development Host.

## Validation

The extension validates:
- Missing tasks in edges
- Cycle detection
- Empty task list

Errors are pushed into VS Code Problems panel.
