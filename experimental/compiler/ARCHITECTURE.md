# FlowForge Compiler Architecture

## Design Overview

The compiler transforms FlowForge IR (Intermediate Representation) into executor-specific artifacts (Argo Workflows YAML, Apache Airflow Python DAGs).

### Architecture Diagram

```
┌─────────────────────┐
│   IR Input          │
│  (PipelineSpec)     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Parser            │  Deserialize IR JSON
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Validator         │  Schema validation
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Optimizer         │  Detect parallelism
│   - Parallelizer    │  Merge sequences
│   - Merger          │  Resource planning
└──────────┬──────────┘
           │
    ┌──────┴──────┐
    │             │
    ▼             ▼
┌────────┐   ┌──────────┐
│  Argo  │   │ Airflow  │  Executor-specific
│Compiler│   │ Compiler │  compilers
└────┬───┘   └────┬─────┘
     │            │
     ▼            ▼
┌─────────┐   ┌──────────┐
│ workflow│   │  DAG     │  Output artifacts
│.yaml    │   │ .py      │
└─────────┘   └──────────┘
```

## Design Decisions

### 1. **Pipeline Architecture**
- **Stage-based**: Parse → Validate → Optimize → Compile
- **Why**: Clear separation of concerns, each stage is independently testable
- **Tradeoff**: More code than single-pass, but easier to extend

### 2. **Executor Abstraction**
- **Interface-based**: ExecutorCompiler interface with ArgoCompiler, AirflowCompiler implementations
- **Why**: Extensible for future executors (Kubernetes, Beam, Spark)
- **Tradeoff**: More abstraction, but enables plugin architecture

### 3. **Optimization Pass**
- **Automatic**: Detect parallelizable tasks, suggest resource configs
- **Why**: Improve output efficiency without manual intervention
- **Tradeoff**: More complex, but provides value

### 4. **Output Validation**
- **Schema-based**: Validate outputs against Argo/Airflow schemas
- **Why**: Fail fast if output is invalid
- **Tradeoff**: Additional dependencies (argo/airflow SDKs)

### 5. **Independence**
- **No SDK dependency**: Compiler depends only on IR module
- **Why**: Compiler can be used standalone, e.g., from CLI or other tools
- **Tradeoff**: Some duplication of schema definitions

## Module Structure

```
compiler/
├── pkg/
│   ├── compiler.go          # Main Compiler interface & factory
│   ├── optimizer.go         # Optimization engine
│   ├── validator.go         # Output validators
│   └── executors/
│       ├── executor.go      # ExecutorCompiler interface
│       ├── argo/            # Argo Workflows compiler
│       │   ├── compiler.go
│       │   └── builder.go
│       └── airflow/         # Apache Airflow compiler
│           ├── compiler.go
│           └── builder.go
├── internal/
│   ├── optimizer/
│   │   ├── parallelizer.go  # Detect fan-out/fan-in
│   │   └── merger.go        # Merge sequential tasks
│   └── validators/
│       ├── argo_validator.go
│       └── airflow_validator.go
├── tests/
│   ├── unit/
│   ├── integration/
│   └── fixtures/
└── examples/
```

## Key Interfaces

### ExecutorCompiler Interface
```go
type ExecutorCompiler interface {
    Compile(ctx context.Context, spec *PipelineSpec) (CompileResult, error)
    Validate(ctx context.Context, result CompileResult) error
    GetFormat() ExecutorFormat
}
```

### Compiler Pipeline
```
IR Input
  ↓ Parse
PipelineSpec (validated)
  ↓ Validate
IR Errors (if any)
  ↓ Optimize
Optimized PipelineSpec
  ↓ Compile (via ExecutorCompiler)
Executor Artifact (YAML/Python)
```

## Tradeoffs

| Decision | Benefit | Tradeoff |
|----------|---------|----------|
| **Stage-based pipeline** | Clear separation, easy testing | More code |
| **Interface abstraction** | Extensible, plugin-ready | Complexity |
| **Automatic optimization** | Better outputs | May not match user intent |
| **Output validation** | Fail fast | Argo/Airflow SDK dependency |
| **IR-only dependency** | Standalone compiler | Schema duplication |

## File Overview

### Core Components

- **compiler.go** (200 lines)
  - Compiler interface with pipeline stages
  - Factory for creating executor-specific compilers
  - Main compile method orchestrating all stages

- **optimizer.go** (150 lines)
  - Parallelizer: detect fan-out/fan-in patterns
  - Merger: merge sequential tasks
  - Resource planner: recommend configs

- **validator.go** (100 lines)
  - Schema validator for outputs
  - Semantic validator for workflow correctness

### Argo Compiler

- **argo/compiler.go** (250 lines)
  - ArgoCompiler implementation
  - Task → ArgoTask transformation
  - Edge → ArgoEdge mapping

- **argo/builder.go** (200 lines)
  - ArgoWorkflowBuilder fluent API
  - YAML serialization
  - Template generation

### Airflow Compiler

- **airflow/compiler.go** (250 lines)
  - AirflowCompiler implementation
  - Task → Operator transformation
  - Edge → DAG edge mapping

- **airflow/builder.go** (200 lines)
  - AirflowDAGBuilder fluent API
  - Python code generation
  - Dependency management

### Tests

- **unit/**: 30+ tests (compiler, optimizer, validators)
- **integration/**: 10+ tests (roundtrip, end-to-end)
- **fixtures/**: Sample IR specs for testing

## Next Steps

1. Create interface definitions (compiler.go, executor.go)
2. Create optimizer implementation
3. Implement Argo compiler
4. Implement Airflow compiler
5. Add comprehensive tests
6. Create example compilations
