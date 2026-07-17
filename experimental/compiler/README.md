# FlowForge Compiler - User Guide

## Overview

The FlowForge Compiler transforms IR (Intermediate Representation) specifications into executor-specific artifacts:
- **Argo Workflows** → YAML workflow definitions
- **Apache Airflow** → Python DAG code

## Installation

```bash
cd compiler
go build -o bin/compiler cmd/compiler/main.go
export PATH=$PATH:$(pwd)/bin
```

## Commands

### compile - Transform IR to executor artifact

```bash
# Compile to Argo Workflows (default)
compiler compile pipeline.json

# Compile to Airflow DAG
compiler compile pipeline.json -executor airflow

# Save to file
compiler compile pipeline.json -output workflow.yaml
compiler compile pipeline.json -executor airflow -output dag.py

# Custom namespace
compiler compile pipeline.json -namespace production
```

### validate - Check IR validity

```bash
# Validate IR specification
compiler validate pipeline.json

# Output: ✓ Pipeline is valid
#         Warnings (if any)
```

### optimize - Apply optimization passes

```bash
# Analyze and suggest optimizations
compiler optimize pipeline.json

# Output: Optimization Summary
#         - Parallelization Detection [APPLIED]
#         - Resource Planning [APPLIED]
```

### inspect - View IR details

```bash
# Display pipeline information
compiler inspect pipeline.json

# Output: Pipeline metadata, tasks, edges, validation status
```

## Compilation Pipeline

Each compilation goes through 5 stages:

```
1. Parse      → Load IR from JSON
2. Validate   → Check semantic validity (cycles, edges, etc.)
3. Optimize   → Detect parallelizable tasks, suggest resources
4. Compile    → Execute executor-specific compiler (Argo/Airflow)
5. Validate   → Verify output format correctness
```

## Examples

### Simple ETL

Input IR:
```json
{
  "metadata": {
    "name": "simple_etl",
    "version": "1.0.0",
    "owner": "data_team"
  },
  "tasks": {
    "extract": {"handler": {"type": "python"}, "config": {"image": "python:3.11"}},
    "transform": {"handler": {"type": "python"}, "config": {"image": "python:3.11"}},
    "load": {"handler": {"type": "python"}, "config": {"image": "python:3.11"}}
  },
  "edges": [
    {"from": "extract", "to": "transform"},
    {"from": "transform", "to": "load"}
  ]
}
```

Compile to Argo:
```bash
compiler compile simple_etl.json -output simple_etl_argo.yaml
```

Output: [Argo Workflow YAML](examples/simple_etl_argo.yaml)

Compile to Airflow:
```bash
compiler compile simple_etl.json -executor airflow -output simple_etl_airflow.py
```

Output: [Airflow DAG Python](examples/simple_etl_airflow.py)

### Fan-Out/Fan-In Pattern

```bash
compiler compile fan_out_fan_in.json -output fan_out_fan_in_argo.yaml
compiler compile fan_out_fan_in.json -executor airflow -output fan_out_fan_in_airflow.py
```

See: [Argo Example](examples/fan_out_fan_in_argo.yaml) and [Airflow Example](examples/fan_out_fan_in_airflow.py)

## Compilation Options

### ExecutorFormat

| Format | Output | Use Case |
|--------|--------|----------|
| `argo` | YAML workflow | Kubernetes-native, enterprise |
| `airflow` | Python DAG | Legacy, hybrid cloud |

### Namespaces

Kubernetes namespace for Argo deployments:
- Default: `default`
- Production: `-namespace production`

### Output Validation

Both Argo and Airflow outputs are validated:
- Argo: YAML structure, required fields (apiVersion, kind, metadata, spec)
- Airflow: Python syntax, airflow imports, DAG definition

## Error Handling

Validation errors prevent compilation:

```bash
$ compiler compile invalid.json
Compilation failed: validation errors: [pipeline must contain at least one task]

$ compiler compile cyclic.json
Compilation failed: validation errors: [pipeline contains a cycle]

$ compiler compile broken_edges.json
Compilation failed: validation errors: [edge references non-existent task: unknown]
```

## Optimization Features

### Parallelization Detection

Identifies tasks that can execute in parallel:
```
Task A can execute in parallel (fan-out, out-degree=2)
Task B waits for multiple tasks (fan-in, in-degree=2)
```

### Resource Planning

Suggests resource configurations based on handler type:
```
Task extract (Python) suggested resources: 1-2 CPU, 512Mi-1Gi memory
Task transform (Python) suggested resources: 1-2 CPU, 512Mi-1Gi memory
```

## IR Format

Required fields in PipelineSpec:
- `metadata.name` - Pipeline name
- `tasks` - Map of task ID → Task definition
- `tasks[*].handler` - Handler type (python, bash, spark)
- `tasks[*].config.image` - Container image
- `edges` - List of task connections

Optional fields:
- `metadata.version` - Pipeline version
- `metadata.owner` - Owner identifier
- `metadata.description` - Description
- `tasks[*].config.env` - Environment variables
- `tasks[*].config.resources` - Resource requests/limits
- `tasks[*].config.timeout` - Task timeout

## Integration with SDK

### From Python SDK

```python
from flowforge import pipeline, task, pipeline_to_ir

@pipeline(name="etl")
def my_pipeline():
    pass

# ... register tasks ...

# Export to IR
spec = pipeline_to_ir(my_pipeline)
spec.to_json()  # Save this to file
```

Then compile with the CLI:
```bash
compiler compile my_pipeline.json -output workflow.yaml
```

## Integration with Orchestrators

### Deploy to Argo

```bash
# Compile
compiler compile pipeline.json -output workflow.yaml

# Deploy
kubectl apply -f workflow.yaml
```

### Deploy to Airflow

```bash
# Compile
compiler compile pipeline.json -executor airflow -output dag.py

# Deploy
cp dag.py $AIRFLOW_HOME/dags/
```

## Performance Considerations

| Operation | Time | Notes |
|-----------|------|-------|
| Parse | < 1ms | JSON deserialization |
| Validate | < 5ms | DAG traversal, cycle detection |
| Optimize | < 10ms | Multiple passes, heuristic analysis |
| Compile (Argo) | < 20ms | Template generation |
| Compile (Airflow) | < 20ms | Python code generation |

## Troubleshooting

### "pipeline must contain at least one task"
- Add at least one task to the IR before compiling

### "edge references non-existent task"
- Check edge `from` and `to` fields match task IDs

### "pipeline contains a cycle"
- Ensure task dependencies form a DAG (no circular references)

### "task has no handler type"
- All tasks must have `handler.type` specified

### "Argo YAML validation failed"
- Check that image names are valid
- Ensure all required metadata fields are present

## Advanced Usage

### Custom Namespaces

```bash
# Dev environment
compiler compile pipeline.json -namespace dev

# Production environment
compiler compile pipeline.json -namespace prod
```

### CI/CD Integration

```bash
# Validate pipeline
if ! compiler validate pipeline.json; then
    echo "Pipeline validation failed"
    exit 1
fi

# Compile to multiple formats
compiler compile pipeline.json -output workflow_argo.yaml
compiler compile pipeline.json -executor airflow -output workflow_airflow.py

# Deploy
kubectl apply -f workflow_argo.yaml
```

## Design Tradeoffs

| Decision | Benefit | Tradeoff |
|----------|---------|----------|
| **Stage-based pipeline** | Clear separation, easy testing | More code |
| **IR input** | Multi-source support | Requires IR generation |
| **Argo support** | Kubernetes-native | YAML complexity |
| **Airflow support** | Legacy systems | Python code complexity |
| **Output validation** | Fail fast | Executor SDK dependency |

## Future Enhancements

- [ ] Conditional branching (if/else compilation)
- [ ] Dynamic task generation (for-each loops)
- [ ] Cost estimation pass
- [ ] Resource auto-tuning
- [ ] Additional executors (Kubernetes Jobs, Apache Beam, Spark)
- [ ] Multi-executor deployments
