# FlowForge Python SDK

> Declarative data pipeline framework for orchestrating workflows across multiple executors.

## Overview

FlowForge Python SDK enables developers to define data pipelines using familiar Python idioms (decorators, type hints) and automatically compile them to multiple executor formats (Argo Workflows, Apache Airflow, local execution).

### Key Features

- **Declarative API**: Use Python decorators to define tasks and pipelines
- **Type-Safe**: Leverage Python type hints for schema inference
- **Multi-Executor**: Compile to Argo, Airflow, or local execution
- **Developer-Friendly**: Fast feedback loop with local execution
- **IR Export**: Convert to FlowForge Intermediate Representation
- **Validation**: Built-in validation before compilation

## Installation

```bash
pip install flowforge
```

## Quick Start

### Define a Pipeline

```python
from flowforge import pipeline, task

@pipeline(name="etl", version="1.0.0")
def my_pipeline():
    """My data pipeline."""
    pass

@task(image="python:3.11")
def extract() -> list:
    """Extract data from source."""
    return [{"id": 1, "name": "Alice"}]

@task(image="python:3.11")
def transform(data: list) -> list:
    """Transform data."""
    return [{"id": r["id"], "name": r["name"].upper()} for r in data]

@task(image="python:3.11")
def load(data: list) -> None:
    """Load to destination."""
    print(f"Loaded {len(data)} records")

# Register tasks
my_pipeline.add_task("extract", extract)
my_pipeline.add_task("transform", transform)
my_pipeline.add_task("load", load)

# Connect tasks
my_pipeline.add_edge(extract, "result", transform, "data")
my_pipeline.add_edge(transform, "result", load, "data")
```

### Execute Locally

```python
from flowforge import LocalExecutor

executor = LocalExecutor()
results = executor.execute(my_pipeline)
```

### Export to IR

```python
from flowforge import pipeline_to_ir

spec = pipeline_to_ir(my_pipeline)
print(spec.to_json())
```

### Use CLI

```bash
# Validate pipeline
flowforge validate pipeline.py

# Compile to IR
flowforge compile pipeline.py -o pipeline.json

# Run locally
flowforge run-local pipeline.py

# Inspect pipeline
flowforge inspect pipeline.py
```

## Core Concepts

### Tasks

Tasks are individual units of work. Define them using the `@task` decorator:

```python
@task(
    image="python:3.11",
    timeout="3600s",
    retries=3,
)
def my_task(input_data: list) -> dict:
    """Task description."""
    return {"processed": len(input_data)}
```

**Decorator Options**:
- `image`: Docker image for execution
- `timeout`: Task timeout (e.g., "3600s")
- `retries`: Number of retries on failure
- `command`: Custom command
- `env`: Environment variables
- `resources`: Resource requests/limits

### Pipelines

Pipelines orchestrate tasks and define their connectivity:

```python
@pipeline(
    name="my_pipeline",
    version="1.0.0",
    owner="data_team",
    description="Description",
)
def my_pipeline():
    pass
```

**Operations**:
- `add_task(id, task)`: Register a task
- `add_edge(from_task, from_port, to_task, to_port)`: Connect tasks
- `validate()`: Check pipeline validity
- `get_tasks()`: Get all tasks
- `get_edges()`: Get all connections

### Schema Inference

Type hints automatically infer JSON schemas:

```python
from typing import List, Dict

@task()
def process(data: List[Dict[str, int]]) -> List[Dict[str, str]]:
    """Schema inferred from type hints."""
    pass
```

**Supported Types**:
- Basic: `int`, `str`, `float`, `bool`
- Containers: `list`, `dict`
- Generics: `List[T]`, `Dict[K, V]`, `Optional[T]`

### Local Execution

Execute pipelines locally for development and testing:

```python
from flowforge import LocalExecutor

executor = LocalExecutor(workers=4, cache_dir="./.cache")
results = executor.execute(pipeline_obj)

# Access results
for task_id, result in results.items():
    print(f"{task_id}: {result}")
```

### IR Export

Convert to FlowForge Intermediate Representation:

```python
from flowforge import pipeline_to_ir

spec = pipeline_to_ir(pipeline_obj)

# Validate
spec.validate()

# Serialize
json_str = spec.to_json()

# Export to file
with open("pipeline.json", "w") as f:
    f.write(json_str)
```

### Visualization

Visualize pipeline structure:

```python
from flowforge import DAGVisualizer

# ASCII visualization
print(DAGVisualizer.visualize(pipeline_obj))

# Graphviz DOT format
dot_code = DAGVisualizer.to_graphviz(pipeline_obj)
```

## Built-in Decorators

Common task decorators are provided:

```python
from flowforge import kafka, s3_read, sql_read, save, s3_write, transform

# Sources
data = kafka("topic_name")
records = s3_read(bucket="my-bucket", key="data.json")
results = sql_read(query="SELECT * FROM table")

# Transforms
cleaned = transform(data, image="clean:v1")
aggregated = aggregate(data, key="category")
filtered = filter_data(data, condition="value > 100")

# Sinks
save(results, path="./output")
s3_write(results, bucket="output-bucket", key="data.json")
sql_write(records, table="results")
```

## Design Patterns

### Linear ETL
```
Extract → Transform → Load
```

### Fan-Out/Fan-In
```
Source → [ProcessA, ProcessB] → Merge
```

### Multi-Branch
```
Extract → {Branch1, Branch2, Branch3} → Combine
```

## Testing

Run tests:

```bash
pytest tests/ -v

# With coverage
pytest tests/ --cov=flowforge
```

## Architecture

### Layers

```
Decorators (@task, @pipeline)
    ↓
Core (Task, Pipeline, TaskReference)
    ↓
Schema (Type hints → JSON Schema)
    ↓
Graph (DAG analysis, topological sort)
    ↓
Compiler (Python → IR PipelineSpec)
    ↓
Executor (Local execution)
    ↓
CLI (Command-line interface)
```

### Modules

- **core**: Task, Pipeline, decorators
- **decorators**: Built-in task decorators (kafka, transform, etc)
- **schema**: Schema inference and validation
- **graph**: DAG visualization and analysis
- **compiler**: IR export (PipelineSpec generation)
- **executor**: Local execution engine
- **cli**: Command-line interface

## Tradeoffs & Design Decisions

| Decision | Why | Tradeoff |
|----------|-----|----------|
| **Decorators** | Pythonic, familiar | Less explicit |
| **Runtime graph gen** | Automatic edges | Slower at import |
| **Type hints for schema** | IDE support, standard | Requires annotations |
| **Local execution** | Fast feedback | Single machine only |
| **IR export** | Multi-executor | Additional abstraction |

## Roadmap

- [ ] Conditional branching (if/else)
- [ ] Dynamic loops (for-each)
- [ ] Custom decorators/plugins
- [ ] Schema registry integration
- [ ] Airflow DAG compilation
- [ ] Argo Workflow compilation
- [ ] Monitoring and observability
- [ ] Cost estimation

## Contributing

Contributions welcome! See [CONTRIBUTING.md](../CONTRIBUTING.md).

## License

Apache License 2.0 - See LICENSE file

## Support

- **Issues**: Report bugs in GitHub Issues
- **Documentation**: See [docs/](../docs/)
- **Examples**: See [examples/](examples/)
