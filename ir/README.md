# FlowForge Intermediate Representation (IR)

> The foundation of multi-executor compatibility: A declarative, executor-agnostic pipeline specification.

## Overview

The **Intermediate Representation (IR)** is the core abstraction layer that enables FlowForge to support multiple executors (Argo Workflows, Apache Airflow, local execution, future Ray/Spark).

### Key Principles

- **Executor Agnostic**: Single IR definition compiles to Argo YAML, Airflow DAG, or local task graph
- **Schema-Driven**: Every task declares input/output schemas using JSON Schema
- **Lineage-Enabled**: Explicit edges enable automatic data lineage tracking
- **Immutable Versioning**: Pipeline specifications are immutable and versioned for replay
- **Extensible Design**: Validators and handlers are pluggable interfaces

---

## Architecture

### IR Data Model

```
PipelineSpec (immutable, versioned)
├── Metadata (name, version, owner, tags)
├── Tasks (map of task ID → Task definition)
│   ├── Task Type (Source, Transform, Sink, Conditional, Retry, Schedule)
│   ├── Handler (what code to run: python, sql, spark, docker, http)
│   ├── Inputs (named ports with JSON schemas)
│   ├── Outputs (named ports with JSON schemas)
│   ├── ExecutorConfig (executor-specific overrides)
│   ├── RetryPolicy (exponential backoff config)
│   └── CostEstimate (optional cost tracking)
└── Edges (connections between task ports)
    └── Edge (from task.port → to task.port)
```

### Task Types

| Type | Purpose | Example |
|------|---------|---------|
| **Source** | Ingest data | Read from S3, database, API |
| **Transform** | Process data | Python, SQL, Spark jobs |
| **Sink** | Export data | Write to S3, database, webhook |
| **Conditional** | Branch logic | if/else based on data |
| **Retry** | Error handling | Wrapper with exponential backoff |
| **Schedule** | Triggering | Cron schedule or event trigger |

---

## Usage

### Go SDK

#### Building a Pipeline

```go
package main

import (
	"fmt"
	"flowforge/ir/pkg"
)

func main() {
	// Create a simple ETL pipeline
	spec, err := ir.NewBuilder("etl-pipeline").
		SetVersion("1.0.0").
		SetOwner("data_team").
		SetDescription("Extract → Transform → Load").
		AddTaskWithDescription(
			"extract",
			ir.TaskTypeSource,
			ir.Handler{Type: "python", Source: "extract_data()"},
			"Extract from S3",
		).
		AddOutput("extract", "records", ir.Schema{"type": "array"}).
		AddTaskWithDescription(
			"transform",
			ir.TaskTypeTransform,
			ir.Handler{Type: "python", Source: "transform_data()"},
			"Transform data",
		).
		AddInput("transform", "data", ir.Schema{"type": "array"}).
		AddOutput("transform", "result", ir.Schema{"type": "array"}).
		AddTaskWithDescription(
			"load",
			ir.TaskTypeSink,
			ir.Handler{Type: "python", Source: "load_data()"},
			"Load to Postgres",
		).
		AddInput("load", "data", ir.Schema{"type": "array"}).
		AddEdge("extract", "records", "transform", "data").
		AddEdge("transform", "result", "load", "data").
		SetExecutorConfig("transform", "argo", map[string]interface{}{
			"image": "python:3.11",
			"resources": map[string]interface{}{
				"cpu": "2", "memory": "4Gi",
			},
		}).
		SetRetryPolicy("transform", &ir.RetryPolicy{
			MaxAttempts:         3,
			Backoff:             "exponential",
			BackoffMultiplier:   2.0,
			InitialDelaySeconds: 5,
		}).
		Build()

	if err != nil {
		panic(err)
	}

	// Serialize to JSON
	data, _ := spec.ToJSON()
	fmt.Println(string(data))
}
```

#### Validating a Pipeline

```go
// Using composite validator
validator := ir.NewCompositeValidator(
	validator.NewDAGValidator(),      // Check for cycles
	validator.NewSchemaValidator(),   // Validate schemas match
)

if err := validator.Validate(spec); err != nil {
	fmt.Println("Validation failed:", err)
}
```

#### Working with Graphs

```go
// Create graph for topological analysis
g := graph.NewDAG(spec)

// Check for cycles
if g.HasCycle() {
	cycle := g.GetCycle()
	fmt.Println("Cycle detected:", cycle)
}

// Get execution order
sorted, err := g.TopologicalSort()
if err != nil {
	panic(err)
}
fmt.Println("Execution order:", sorted)

// Check dependencies
predecessors := g.Predecessors("transform")
fmt.Println("Tasks that feed into transform:", predecessors)
```

---

### Python SDK

#### Building a Pipeline

```python
from flowforge.ir import pipeline, TaskType, Handler, RetryPolicy, CostEstimate, CostDimension

# Create using builder
spec = (
    pipeline("etl-pipeline")
    .set_version("1.0.0")
    .set_owner("data_team")
    .set_description("Extract → Transform → Load")
    .add_task(
        "extract",
        TaskType.SOURCE,
        Handler(type="python", source="extract_data()"),
        description="Extract from S3",
    )
    .add_output("extract", "records", {"type": "array"})
    .add_task(
        "transform",
        TaskType.TRANSFORM,
        Handler(type="python", source="transform_data()"),
        description="Transform data",
    )
    .add_input("transform", "data", {"type": "array"})
    .add_output("transform", "result", {"type": "array"})
    .add_task(
        "load",
        TaskType.SINK,
        Handler(type="python", source="load_data()"),
        description="Load to Postgres",
    )
    .add_input("load", "data", {"type": "array"})
    .add_edge("extract", "records", "transform", "data")
    .add_edge("transform", "result", "load", "data")
    .set_executor_config("transform", "argo", {
        "image": "python:3.11",
        "resources": {"cpu": "2", "memory": "4Gi"},
    })
    .set_retry_policy("transform", RetryPolicy(
        max_attempts=3,
        backoff="exponential",
        backoff_multiplier=2.0,
        initial_delay_seconds=5,
    ))
    .build()

# Serialize to JSON
print(spec.to_json())
```

#### Validating a Pipeline

```python
from flowforge.ir import DAGValidator, SchemaValidator, CompositeValidator

# Validate with composite validator
validator = CompositeValidator([
    DAGValidator(),        # Check for cycles
    SchemaValidator(),     # Validate schemas
])

error = validator.validate(spec)
if error:
    print(f"Validation failed: {error}")
```

#### Working with Graphs

```python
from flowforge.ir import DAGGraph

# Create graph
graph = DAGGraph(spec)

# Check for cycles
if graph.has_cycle():
    cycle = graph.get_cycle()
    print(f"Cycle: {cycle}")

# Get execution order
sorted_tasks, error = graph.topological_sort()
if error:
    print(f"Error: {error}")
else:
    print(f"Execution order: {sorted_tasks}")

# Check dependencies
predecessors = graph.predecessors("transform")
print(f"Tasks feeding into transform: {predecessors}")
```

---

## JSON Schema Specification

The IR is serialized using a standard JSON Schema. See [spec.json](spec.json) for the complete schema.

### Example Pipeline (JSON)

```json
{
  "apiVersion": "flowforge.io/v1",
  "kind": "Pipeline",
  "metadata": {
    "name": "etl-pipeline",
    "version": "1.0.0",
    "owner": "data_team"
  },
  "tasks": {
    "extract": {
      "type": "Source",
      "handler": {"type": "python", "source": "extract_data()"},
      "outputs": {"records": {"type": "array"}}
    },
    "transform": {
      "type": "Transform",
      "handler": {"type": "python", "source": "transform_data()"},
      "inputs": {"data": {"type": "array"}},
      "outputs": {"result": {"type": "array"}},
      "retry": {
        "maxAttempts": 3,
        "backoff": "exponential",
        "backoffMultiplier": 2.0,
        "initialDelaySeconds": 5
      }
    },
    "load": {
      "type": "Sink",
      "handler": {"type": "python", "source": "load_data()"},
      "inputs": {"data": {"type": "array"}}
    }
  },
  "edges": [
    {"from": {"task": "extract", "port": "records"}, "to": {"task": "transform", "port": "data"}},
    {"from": {"task": "transform", "port": "result"}, "to": {"task": "load", "port": "data"}}
  ]
}
```

See [examples/](examples/) for more pipeline examples.

---

## Design Decisions

### 1. Task-Centric DAG Model

✅ **Why**: Both Argo and Airflow use DAGs as their fundamental model  
✅ **Enables**: Multi-executor compilation from single IR  
⚠️ **Tradeoff**: Less expressive than Argo nested workflows

### 2. Schema-Driven Execution

✅ **Why**: Standard format (JSON Schema) enables validation and tooling  
✅ **Enables**: Type safety, lineage tracking, cost estimation  
⚠️ **Tradeoff**: Verbose for complex types

### 3. Executor-Specific Configuration

✅ **Why**: Different executors have different configuration needs  
✅ **Enables**: Argo resource specs, Airflow pools, etc.  
⚠️ **Tradeoff**: More complex structure

### 4. Immutable Specifications

✅ **Why**: Enables versioning, caching, audit trails, replay  
✅ **Enables**: Historical analysis, cost tracking over time  
⚠️ **Tradeoff**: Need separate execution context for mutations

### 5. Explicit Edges for Lineage

✅ **Why**: Clear data flow enables automatic lineage tracking  
✅ **Enables**: Data governance, PII detection, cost attribution  
⚠️ **Tradeoff**: More verbose than implicit dependencies

### 6. Pluggable Validators

✅ **Why**: Extensible validation without modifying core  
✅ **Enables**: Custom rules per organization  
⚠️ **Tradeoff**: More complexity than monolithic validator

---

## Validation Rules

### DAG Validation
- ✅ No cycles allowed (must be acyclic DAG)
- ✅ All edges reference valid tasks
- ✅ All ports must exist on referenced tasks

### Schema Validation
- ✅ Input ports must have matching input schemas
- ✅ Output ports must match edge expectations

### Resource Validation (future)
- ✅ Sufficient memory/CPU for task type
- ✅ Executor supports handler type

---

## Extensibility

### Adding a New Handler Type

```go
// 1. Define handler in JSON schema
"handler": {
  "type": "string",
  "enum": ["python", "sql", "spark", "docker", "http", "my_custom_handler"]
}

// 2. Implement compilation in compiler/ module
type MyCustomCodeGenerator struct{}

func (cg *MyCustomCodeGenerator) Generate(task *ir.Task) (string, error) {
    // Generate executor-specific code
    return code, nil
}

// 3. Register in compiler registry
registry.Register("my_custom_handler", &MyCustomCodeGenerator{})
```

### Adding a New Validator

```go
// 1. Implement Validator interface
type CustomValidator struct {}

func (cv *CustomValidator) Validate(spec *ir.PipelineSpec) error {
    // Custom validation logic
    return nil
}

func (cv *CustomValidator) Name() string {
    return "CustomValidator"
}

// 2. Use in composite validator
validator := ir.NewCompositeValidator(
    validator.NewDAGValidator(),
    &CustomValidator{},
)
```

---

## Public API (pkg/)

### Go Interfaces

- **`Builder`** - Fluent API for constructing pipelines
- **`Validator`** - Interface for validation plugins
- **`Graph`** - Graph operations (topological sort, cycle detection)

### Python Classes

- **`PipelineBuilder`** - Fluent API for constructing pipelines
- **`Validator`** - Base class for validation plugins
- **`DAGGraph`** - Graph operations

All exported via:
- Go: `flowforge/ir/pkg`
- Python: `flowforge.ir`

---

## Testing

### Go Unit Tests

```bash
cd ir/tests/unit
go test -v

# Coverage
go test -cover ./...
```

### Python Unit Tests

```bash
cd sdk
pytest flowforge/ir/tests/ -v

# Coverage
pytest --cov=flowforge.ir flowforge/ir/tests/
```

### Test Examples

See [tests/unit/](tests/unit/) for comprehensive test coverage:
- `spec_test.go` - Pipeline spec creation, validation
- `validator_test.go` - DAG and schema validation
- `graph_test.go` - Topological sort, cycle detection

---

## Examples

The [examples/](examples/) directory contains sample pipelines:

1. **simple_etl.json** - Extract → Transform → Load (3-task linear)
2. **fan_out_fan_in.json** - Parallel processing (fan-out/fan-in)
3. **data_quality.json** - Data quality checks with conditional branching
4. **scheduled_batch.json** - Scheduled batch job with retries

Load and validate:

```go
data, _ := ioutil.ReadFile("examples/simple_etl.json")
spec, err := ir.FromJSON(data)
```

```python
import json
with open("examples/simple_etl.json") as f:
    spec = PipelineSpec.from_dict(json.load(f))
```

---

## Future Enhancements

- **Dynamic Tasks** - Conditional loops (repeat until condition)
- **Custom Handlers** - Plugin system for user-defined handlers
- **Schema Inference** - Auto-infer schemas from code analysis
- **Cost Optimization** - Optimize task placement for cost
- **Lineage Export** - Export lineage to data catalogs (OpenMetadata, Collibra)
- **Version Migration** - Handle schema evolution gracefully

---

## Integration with Other Modules

### Compiler

IR is **input** to compiler:
```
PipelineSpec (IR) → Compiler → [ArgoWorkflow, AirflowDAG, LocalDAG]
```

The compiler uses IR interfaces to generate executor-specific code.

### Storage

IR is **persisted** to storage:
```
PipelineSpec → Storage (PostgreSQL) → Retrieved for execution/replay
```

### Observability

IR enables **observability**:
```
Edges → Lineage tracking → Metrics → Tracing
```

### SDK

SDK uses IR builder to construct pipelines from Python.

---

## Performance Notes

- **Serialization**: JSON serialization takes ~1ms for 100-task pipeline
- **Validation**: DAG validation O(V+E) where V=tasks, E=edges
- **Topological Sort**: O(V+E) using Kahn's algorithm
- **Memory**: Typical pipeline ~10KB JSON, 1MB Go struct

---

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for contribution guidelines.

### Module Boundaries

- **Public**: `pkg/*.go` (interfaces and types)
- **Private**: `internal/*/*.go` (implementations)
- **Tests**: `tests/unit/*.go`, `tests/integration/*.go`
- **Examples**: `examples/*.json`

Imports must follow:
- ✅ Other modules can import `ir/pkg`
- ❌ Other modules cannot import `ir/internal`

---

## License

Apache License 2.0 - See LICENSE file

---

## Support

- **Issues**: Report bugs in GitHub Issues
- **Documentation**: See [docs/](../docs/) for user/developer guides
- **Questions**: Use GitHub Discussions or Slack

---

## Related Documentation

- [ARCHITECTURE.md](../ARCHITECTURE.md) - System design
- [compiler/README.md](../compiler/README.md) - IR compilation
- [sdk/README.md](../sdk/README.md) - Python SDK usage
- [examples/](examples/) - Sample pipelines
