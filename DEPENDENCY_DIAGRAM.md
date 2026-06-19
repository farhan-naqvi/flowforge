# Module Responsibilities & Dependency Diagram

## 1. Detailed Module Responsibilities Matrix

| Module | Purpose | Owners | Tech Stack | Key Interfaces | Deployment |
|--------|---------|--------|-----------|-----------------|------------|
| **IR** | Core data structures, validation rules, builder | Compiler team | Go | `PipelineSpec`, `TaskSpec`, `ValidationError` | Embedded in `api/`, `compiler/` |
| **Parser** | Convert SDK/YAML/UI events → AST | SDK + UI teams | Go, Python, TypeScript | `Parser`, `ParseResult`, `ParseError` | Distributed (`api/compiler`) |
| **Compiler** | AST → executor config (Argo/Airflow/Local) | Compiler team | Go | `Compiler`, `OptimizationPass`, `CodeGenerator` | `compiler/` service |
| **Argo Driver** | Submit/monitor workflows on K8s | Executor team | Go | `ExecutorDriver`, `SubmissionRequest`, `Status` | `executor/` service |
| **Local Driver** | In-process task execution (dev mode) | Executor team | Go, Python | `ExecutorDriver` | `executor/` service |
| **Airflow Driver** | DAG generation + scheduler integration | Executor team | Python, Go | `ExecutorDriver`, `DAGSpec` | `executor/` service (future) |
| **Lineage Engine** | Track data provenance, build graph | Core team | Go, PostgreSQL | `LineageStore`, `LineageQuery`, `Graph` | `compiler/` service |
| **Cost Engine** | Estimate/track resource costs | Core team | Go, PostgreSQL | `CostCalculator`, `CostEstimate` | `compiler/` service |
| **API Server** | gRPC + REST entry point | Platform team | Go | gRPC services, REST routes | `api/` service |
| **Database Layer** | PostgreSQL access, migrations | Core team | Go, SQL | `Store`, `Query`, `Transaction` | `api/` service |
| **Auth** | OIDC, RBAC, API token validation | Platform team | Go | `AuthMiddleware`, `AccessControl` | `api/` service |
| **Observability** | Metrics, logging, tracing | Platform team | Go, OTEL, Prometheus | `MetricsCollector`, `Logger`, `Tracer` | `api/` service |
| **Python SDK** | User-facing Pipeline API, CLI | SDK team | Python | `Pipeline`, `Task`, `Client` | Standalone package |
| **UI** | Visual builder, explorer, dashboard | UI team | React, TypeScript | REST clients, component hierarchy | Standalone SPA |
| **Helm Charts** | K8s resource manifests | DevOps team | Helm, Kubernetes YAML | Chart values, templates | Separate deployment |
| **Terraform** | Infrastructure provisioning | DevOps team | Terraform, HCL | Modules, variables, outputs | Separate deployment |

---

## 2. Dependency Diagram

### 2.1 High-Level Module Dependencies

```
                    ┌─────────────┐
                    │ Python SDK  │
                    │ + CLI       │
                    └──────┬──────┘
                           │ gRPC
                    ┌──────▼──────┐
                    │  API Server │
                    │   (Go)      │
                    └──────┬──────┘
                           │
            ┌──────────────┼──────────────┐
            │              │              │
      ┌─────▼────┐  ┌─────▼────┐  ┌─────▼────┐
      │ Compiler │  │    DB    │  │   Auth   │
      │  (Go)    │  │ (Postgres)│  │  (OIDC)  │
      └─────┬────┘  └──────────┘  └──────────┘
            │
      ┌─────▼─────────────────┐
      │  Compiler Internals   │
      │  ┌───────────────┐    │
      │  │ Parser Module │    │
      │  │ (YAML/SDK/UI) │    │
      │  └───────────────┘    │
      │  ┌───────────────┐    │
      │  │ Optimizer     │    │
      │  └───────────────┘    │
      │  ┌───────────────┐    │
      │  │ Codegen       │    │
      │  │ (Argo/AF/Loc) │    │
      │  └───────────────┘    │
      │  ┌───────────────┐    │
      │  │ Lineage Eng.  │    │
      │  │ + Cost Eng.   │    │
      │  └───────────────┘    │
      └─────┬─────────────────┘
            │ ExecutorDriver interface
            │
      ┌─────▼─────────────────────────────┐
      │      Executor Layer                │
      │  ┌──────────┐ ┌──────────┐       │
      │  │ Argo     │ │ Local    │       │
      │  │ Driver   │ │ Driver   │       │
      │  └──────────┘ └──────────┘       │
      │  ┌──────────────────────────┐    │
      │  │ Airflow Driver (future)  │    │
      │  └──────────────────────────┘    │
      └─────────────────────────────────────┘
            │
      ┌─────▼──────────────────┐
      │  Runtime               │
      │  ┌────────────────┐    │
      │  │ Kubernetes +   │    │
      │  │ Argo Workflows │    │
      │  └────────────────┘    │
      │  ┌────────────────┐    │
      │  │ Local Python   │    │
      │  │ (dev mode)     │    │
      │  └────────────────┘    │
      │  ┌────────────────┐    │
      │  │ Airflow        │    │
      │  │ (future)       │    │
      │  └────────────────┘    │
      └────────────────────────┘
```

### 2.2 Detailed Compiler Dependencies

```
IR Spec (immutable)
    ▲
    │
    └─── Parser ────────────┬──── SDK Parser
         (Interface)        ├──── YAML Parser
                            └──── UI Parser (events)
                                
    ▲
    │
    ├─── Validator
    │    ├── Schema checks
    │    ├── DAG cycle detection
    │    └── Resource constraint validation
    │
    ├─── Optimizer (Optional passes)
    │    ├── Task merging
    │    ├── Parallelism maximization
    │    └── Resource pooling
    │
    ├─── CodeGenerator (Executor-specific)
    │    ├── Argo Workflow YAML
    │    ├── Airflow DAG Python
    │    └── Local DAG (Go struct)
    │
    ├─── Lineage Engine
    │    ├── Build provenance graph
    │    └── Store in PostgreSQL
    │
    └─── Cost Engine
         ├── Estimate task costs
         └── Store projections in PostgreSQL
```

### 2.3 Runtime Execution Flow

```
Execution Request (from API)
    │
    ├─ Fetch PipelineSpec from DB
    ├─ Invoke Compiler
    │   └─ Output: ExecutorConfig (Argo/Local/Airflow)
    │
    ├─ Invoke ExecutorDriver (based on config.executor_type)
    │   │
    │   ├─ Argo Driver:
    │   │   └─ Submit Argo Workflow YAML to K8s API
    │   │       └─ Kubernetes runs pods
    │   │
    │   ├─ Local Driver:
    │   │   └─ Run tasks in subprocess pool
    │   │       └─ Python execution
    │   │
    │   └─ Airflow Driver (future):
    │       └─ Submit DAG to Airflow Webserver
    │           └─ Airflow Scheduler executes
    │
    ├─ Monitor execution (via driver)
    │   └─ Update status in DB
    │
    ├─ Record lineage (as tasks complete)
    │   └─ Store in PostgreSQL
    │
    └─ Track costs (per task)
        └─ Aggregated dashboard
```

---

## 3. Key Dependency Constraints

### 3.1 Acyclic Dependencies

**Allowed**:
- `API Server` → `Compiler` (uses it)
- `Compiler` → `IR` (consumes IR)
- `Parser` → `IR` (produces IR)
- `CodeGenerator` → `IR` (reads IR)

**Forbidden**:
- Circular: `IR` ←→ `Parser`
- Circular: `Compiler` ←→ `API Server`

**Solution**: Use dependency inversion (interfaces) to break cycles.

### 3.2 Horizontal Scalability

**Stateless modules** (can scale horizontally):
- API Server
- Compiler service
- Executor drivers

**Stateful modules** (require careful scaling):
- PostgreSQL (managed externally)
- Redis cache (cluster mode)
- Prometheus (remote storage)

### 3.3 Independent Deployability

Each module should be independently deployable:

```
flowforge-api:v1.2.0
flowforge-compiler:v1.2.0
flowforge-executor:v1.2.0
flowforge-ui:v1.2.0
```

Achieved via:
- Clear API boundaries (gRPC/REST)
- Semantic versioning
- Backward compatibility window

---

## 4. Interface Contracts (Key Abstractions)

### 4.1 Parser Interface

```go
type Parser interface {
    // Parse input into IR
    Parse(ctx context.Context, input interface{}) (*ir.PipelineSpec, error)
    // Format determines parser type (yaml, sdk, ui)
    Supports(format string) bool
}
```

**Implementations**:
- `YAMLParser` (YAML files)
- `SDKParser` (Python SDK AST)
- `UIParser` (UI builder events)

### 4.2 ExecutorDriver Interface

```go
type ExecutorDriver interface {
    // Submit compiled config to executor
    Submit(ctx context.Context, config *ExecutorConfig) (*Submission, error)
    // Monitor execution status
    Status(ctx context.Context, submissionID string) (*ExecutionStatus, error)
    // Fetch task logs
    Logs(ctx context.Context, submissionID, taskID string) (io.Reader, error)
    // Stop execution
    Cancel(ctx context.Context, submissionID string) error
}
```

**Implementations**:
- `ArgoDriver` (Kubernetes + Argo Workflows)
- `LocalDriver` (in-process Python)
- `AirflowDriver` (Airflow Webserver API)

### 4.3 Optimizer Pass Interface

```go
type OptimizationPass interface {
    // Transform IR → optimized IR
    Optimize(ctx context.Context, pipeline *ir.PipelineSpec) (*ir.PipelineSpec, error)
    // Name of optimization (for logging)
    Name() string
    // Enable/disable per-executor
    AppliesTo(executor string) bool
}
```

**Implementations**:
- `MergeTasksPass` (combine sequential tasks)
- `ParallelizePass` (detect parallelizable tasks)
- `ResourcePoolPass` (share resources across tasks)

### 4.4 CodeGenerator Interface

```go
type CodeGenerator interface {
    // Generate executor-specific config from IR
    Generate(ctx context.Context, pipeline *ir.PipelineSpec) (interface{}, error)
    // Executor type (argo, airflow, local)
    ExecutorType() string
}
```

**Implementations**:
- `ArgoGenerator` (Argo Workflow YAML)
- `AirflowGenerator` (Airflow DAG Python)
- `LocalGenerator` (Go DAG struct)

---

## 5. Cross-Cutting Concerns

### 5.1 Observability Integration

All modules emit:
- **Metrics**: via Prometheus client library
- **Logs**: structured JSON (timestamp, level, module, message)
- **Traces**: OpenTelemetry spans

Example:
```go
// In Compiler
span := tracer.Start(ctx, "compiler.optimize")
defer span.End()
metrics.CompilerDuration.Observe(time.Since(start).Seconds())
```

### 5.2 Error Handling

Consistent error model:
```go
type Error struct {
    Code    ErrorCode   // e.g., VALIDATION_ERROR, SUBMISSION_FAILED
    Message string
    Details map[string]interface{}
    Cause   error       // root cause
}
```

All modules return `Error` (not generic `error`).

### 5.3 Versioning

API compatibility:
- gRPC: semantic versioning in proto package (`flowforge.v1.*`)
- REST: versioning in URL path (`/api/v1/*`)
- Database: migrations tracked with version

---

## 6. Build & Integration Points

### 6.1 Build Order

**Day 1 Builds**:
1. IR + Validator (core)
2. Parser (YAML, SDK stub)
3. Compiler core + Argo codegen
4. Argo Driver
5. API Server (stubs)
6. Python SDK (basic)

**Week 2**:
7. UI builder (React)
8. Local driver
9. Lineage engine
10. Cost engine

**Week 3+**:
11. Advanced features (replay, diff, templates)
12. Airflow driver
13. Self-healing, benchmarking

### 6.2 Integration Testing Strategy

**Layer 1**: Unit tests (each module)
```bash
cd compiler && go test ./...
cd sdk && pytest tests/unit
cd ui && npm test
```

**Layer 2**: Integration tests (module pairs)
```bash
tests/e2e/test_yaml_to_argo.py       # YAML → Argo
tests/e2e/test_sdk_to_local.py       # SDK → Local
tests/e2e/test_lineage_tracking.py   # Execution → Lineage
```

**Layer 3**: E2E tests (full flow)
```bash
# User creates pipeline via SDK
# Submits to API
# Compiler generates Argo YAML
# Argo executes on K8s
# Results in DB + Lineage graph
```

---

## 7. Technology Choices Rationale

| Component | Tech | Why | Alternative | Trade-off |
|-----------|------|-----|-------------|-----------|
| Control plane language | Go | Concurrency, performance, gRPC native | Python | Learning curve for data engineers |
| Serialization | Protobuf | Versioning, cross-language, gRPC native | JSON/YAML | Binary format less readable |
| Storage | PostgreSQL | ACID, JSON support, mature | MongoDB | Schema structure better with SQL |
| Executor 1 | Argo | K8s native, serverless-friendly | Kubernetes Job | Workflow-specific features |
| Executor 2 | Airflow | Enterprise standard, existing integrations | Custom | Different scheduling model |
| Python runtime | containerized | Isolation, reproducibility | In-process | Container overhead |
| UI framework | React | Ecosystem, component libraries | Vue/Svelte | Learning curve for new team |
| IaC | Terraform + Helm | Multi-cloud, declarative | CloudFormation | AWS-specific lock-in |

