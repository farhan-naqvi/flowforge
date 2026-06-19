# FlowForge Architecture

## 1. System Overview

FlowForge is a declarative data pipeline orchestration platform that abstracts away executor specifics through an Intermediate Representation (IR), enabling users to define pipelines once and execute on multiple runtimes (Argo Workflows, Airflow DAGs, local execution).

### Core Vision
- **Abstraction Layer**: Single source of truth (IR) that compiles to multiple executors
- **Multi-Input Support**: Python SDK, YAML, Visual UI → Same IR
- **Multi-Executor Support**: Argo Workflows, Airflow, Local, Future runtimes
- **Enterprise Features**: Lineage, replay, cost visibility, schema contracts, self-healing
- **Not an Airflow Replacement**: Complement, not compete - designed to be Airflow-native where it makes sense

---

## 2. Design Decisions & Rationale

### 2.1 Intermediate Representation (IR)

**Decision**: Define IR as immutable, versioned AST with:
- Task definitions (name, image, command, resources, retries, timeout)
- Dependencies (DAG edges)
- Data contracts (input/output schemas)
- Execution policies (parallelism, retry strategy, cost constraints)
- Lineage metadata (provenance tracking)

**Rationale**:
- Language-agnostic: supports any input format
- Compiler target: multiple backends from single IR
- Auditable: immutable record for compliance/replay
- Extensible: new executors without SDK/UI changes
- Schema contracts: catch data issues early

**Trade-off**: Requires upfront domain modeling; not as flexible as DSL-only approach, but gains portability.

### 2.2 Compiler Architecture

**Decision**: Multi-stage compiler:
1. **Parse** (SDK/YAML/UI → AST)
2. **Validate** (schema, DAG cycles, resource constraints)
3. **Optimize** (merge tasks, optimize parallelism)
4. **Codegen** (Argo Workflow YAML, Airflow DAG Python, Local DAG)

**Rationale**:
- Separation of concerns
- Pluggable validation rules
- Future: custom optimizations per executor
- Reusable components for different backends

**Trade-off**: Complexity upfront, but enables sophisticated features like cost estimation.

### 2.3 Execution Model

**Decision**: Three execution modes (MVP → Argo; Future → Airflow):

| Mode | Use Case | Executor |
|------|----------|----------|
| **Local** | Dev/debugging | Python process pool |
| **Argo** | Kubernetes-native, cloud | Argo Workflows |
| **Airflow** | Enterprise integration | Airflow DAGs |

Execution engine abstraction: plug in any executor via driver interface.

**Rationale**:
- Local mode catches issues early (developer velocity)
- Argo: cloud-native, serverless-friendly
- Airflow: existing enterprise infrastructure
- Driver pattern: executor-agnostic core

**Trade-off**: Must maintain driver contracts; executor-specific features require mapping layer.

### 2.4 Data Contracts & Lineage

**Decision**: 
- Input/output schemas (via Pydantic or JSON Schema)
- Lineage graph: task → output, input ← task
- Runtime validation: schema enforcement at task boundaries
- Lineage storage: PostgreSQL + queryable API

**Rationale**:
- Schema contracts: prevent downstream breakage
- Lineage: data governance, cost attribution, debugging
- Runtime validation: catch schema drift early

**Trade-off**: Runtime overhead (mitigation: optional/async validation).

### 2.5 Backend: Go vs Python

**Decision**: Go backend (not Python)
- Control plane/orchestration: Go
- Python SDK/CLI: wraps Go binary over gRPC
- Transformations: Containerized Python

**Rationale**:
- Go: high concurrency, minimal overhead for gRPC/REST servers
- Python SDK: ease of use for data engineers
- Separation: control plane stability from transformation logic
- Docker: transformations can be any language

**Trade-off**: Polyglot complexity; mitigated by clear service boundaries.

### 2.6 Storage: PostgreSQL

**Decision**: PostgreSQL for:
- Pipeline definitions
- Execution history
- Lineage graph
- User/org/auth data
- Cost metrics

**Rationale**:
- ACID compliance: execution state integrity
- JSON support: flexible schema storage
- Time-series queries: cost/performance analytics
- Mature ecosystem: PostgRES extensions for complex queries

**Alternative Considered**: Separate time-series DB (InfluxDB/VictoriaMetrics) for metrics; decision: start with PostgreSQL, extract later if needed.

---

## 3. Module Architecture

### 3.1 High-Level Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                        UI Layer (React)                         │
│  - Builder (drag-drop pipeline editor)                          │
│  - Explorer (lineage, replay, history)                          │
│  - Dashboard (executions, costs, metrics)                       │
└─────────────────────────────────────────────────────────────────┘
                              ↓ REST/GraphQL
┌─────────────────────────────────────────────────────────────────┐
│                   API Layer (Go, gRPC/REST)                     │
│  - Pipeline API (create, update, list)                          │
│  - Execution API (trigger, status, logs)                        │
│  - Lineage API (query graph)                                    │
│  - Cost/Metrics API                                             │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    Core Engine (Go)                             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Compiler                                                │   │
│  │  - Parser (SDK/YAML/UI events)                          │   │
│  │  - Validator                                            │   │
│  │  - Optimizer                                            │   │
│  │  - Codegen (Argo, Airflow, Local)                       │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              ↓                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Executor Driver (interface)                             │   │
│  │  - Local Driver                                         │   │
│  │  - Argo Driver                                          │   │
│  │  - Airflow Driver (future)                              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              ↓                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ State & Persistence                                     │   │
│  │  - Pipeline Store                                       │   │
│  │  - Execution Store                                      │   │
│  │  - Lineage Store                                        │   │
│  │  - Metrics Store                                        │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                  Runtime Layer                                  │
│  - Local Python runtime                                         │
│  - Kubernetes (Argo Workflows)                                  │
│  - Airflow Scheduler/Executor                                   │
│  - Container Registry (Docker)                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. Module Responsibilities

| Module | Responsibility | Language | Key Interfaces |
|--------|-----------------|----------|-----------------|
| **IR** | AST, schema, validation | Go | `PipelineSpec`, `TaskSpec`, `ValidationRule` |
| **Parser** | SDK/YAML/UI → AST | Go, Python | `Parser` trait, `ParseResult` |
| **Compiler** | AST → Executor config | Go | `Compiler`, `OptimizationPass` |
| **Argo Driver** | Argo submission, monitoring | Go | `ExecutorDriver`, `SubmissionRequest` |
| **Local Driver** | Python process executor | Go/Python | `ExecutorDriver` |
| **Airflow Driver** | Airflow DAG generation | Python/Go | `ExecutorDriver`, `DAGSpec` |
| **Lineage Engine** | Graph storage/query | Go, PostgreSQL | `LineageStore`, `LineageQuery` |
| **Cost Engine** | Resource tracking, estimation | Go | `CostCalculator`, `ResourceUsage` |
| **API Server** | gRPC + REST endpoints | Go | gRPC services, OpenAPI |
| **Python SDK** | High-level user API | Python | `Pipeline`, `Task`, `IRBuilder` |
| **UI** | Builder, explorer, dashboard | React/TypeScript | REST API clients |
| **Helm Charts** | K8s deployment | Helm/Terraform | Values, templates |
| **Tests** | Unit, integration, e2e | Go, Python, Jest | Test fixtures, factories |

---

## 5. Intermediate Representation (IR) Structure

```protobuf
// Simplified proto definition (actual will be more detailed)

message PipelineSpec {
  string id = 1;
  string name = 2;
  string version = 3;
  repeated TaskSpec tasks = 4;
  repeated Edge edges = 5;
  PipelinePolicy policy = 6;
  map<string, SchemaContract> contracts = 7;
}

message TaskSpec {
  string id = 1;
  string name = 2;
  string image = 3;
  repeated string command = 4;
  repeated string args = 5;
  map<string, string> env = 6;
  ResourceRequirement resources = 7;
  RetryPolicy retry = 8;
  string timeout = 9;
  map<string, InputBinding> inputs = 10;
  map<string, OutputBinding> outputs = 11;
}

message Edge {
  string from_task = 1;
  string to_task = 2;
  string output_name = 3;
  string input_name = 4;
}

message SchemaContract {
  string format = 1; // "json-schema" or "pydantic"
  bytes schema = 2;
  bool enforce_at_runtime = 3;
}

message PipelinePolicy {
  int32 max_parallelism = 1;
  string execution_mode = 2; // "local", "argo", "airflow"
  CostPolicy cost_policy = 3;
}
```

---

## 6. Data Flow Examples

### 6.1 Pipeline Definition → Execution (Argo)

```
User (Python SDK)
  ↓
  pipeline = Pipeline("my-pipeline")
  pipeline.add_task("extract", image="postgres-client", ...)
  pipeline.add_task("transform", image="custom-python:1.0", ...)
  pipeline.connect("extract", "transform")
  pipeline.submit()
  ↓
Python SDK (gRPC call)
  ↓
API Server (Go)
  ↓
Parser (SDK → AST)
  ↓
Compiler (AST → Argo YAML)
  ↓
Argo Driver (Submit to K8s API)
  ↓
Kubernetes / Argo Workflows
  ↓
Task Pods (Container execution)
  ↓
State Store (execution history)
  ↓
Lineage Engine (record provenance)
  ↓
Observability (Prometheus)
```

### 6.2 Local Development Mode

```
User (CLI)
  ↓
  ff local pipeline.yaml
  ↓
Parser (YAML → AST)
  ↓
Compiler (AST → Local DAG)
  ↓
Local Driver (in-process execution)
  ↓
Python subprocesses / threads
  ↓
Output files in ./output
```

---

## 7. Deployment Model

### 7.1 Architecture Components

**Control Plane (Go Services)**:
- API Server (gRPC + REST)
- Compiler Service
- State Manager (PostgreSQL)
- Lineage Engine
- Cost Estimator

**Data Plane**:
- Argo Workflows (Kubernetes)
- Airflow (optional, separate cluster)
- Local runner (dev only)

**Supporting Services**:
- PostgreSQL (state)
- Redis (caching, job queue)
- Prometheus (metrics)
- Grafana (dashboards)

### 7.2 Deployment Strategies

**Single Cluster** (MVP):
- Control plane + Argo in same K8s cluster
- PostgreSQL managed (RDS/Cloud SQL)

**Multi-Cluster** (Future):
- Control plane: shared cluster
- Execution: dedicated per environment (dev/staging/prod)

---

## 8. Observability Strategy

### 8.1 Metrics

**System Level**:
- API latency, error rates (gRPC/REST)
- Compiler duration (parse, optimize, codegen)
- Execution queue depth

**Pipeline Level**:
- Task success/failure rate
- Task duration, cost
- Data volume, lineage depth

### 8.2 Logging

**Structured Logs**:
- Execution logs (stderr, stdout from tasks)
- API logs (request/response, errors)
- Compiler logs (optimization decisions)

**Log Storage**:
- PostgreSQL (queryable history)
- Loki (for streaming logs from tasks)

### 8.3 Tracing

**Distributed Tracing**:
- OpenTelemetry instrumentation
- Jaeger backend
- Trace: user request → API → compiler → executor submission

---

## 9. Security Model

### 9.1 Authentication & Authorization

- OIDC integration (GitHub, Google, corporate SSO)
- RBAC: viewer, editor, admin per pipeline/org
- API tokens for programmatic access

### 9.2 Data Protection

- Encryption at rest (PostgreSQL)
- Encryption in transit (TLS for all gRPC/REST)
- Secret management: HashiCorp Vault or cloud provider (AWS Secrets Manager)

### 9.3 Pipeline Security

- Container image signing (Cosign)
- Network policies (K8s NetworkPolicy)
- Resource limits (CPU, memory, timeout)

---

## 10. Testing Strategy

### 10.1 Unit Tests

- IR validation rules
- Compiler passes (optimization, codegen)
- Schema contract validation
- Cost calculations

### 10.2 Integration Tests

- API Server ↔ State Store
- Compiler ↔ Executor Drivers
- Lineage Engine ↔ PostgreSQL

### 10.3 End-to-End Tests

- Python SDK → API → Compiler → Local Driver
- YAML → Argo submission → pod execution
- UI builder → IR → execution

### 10.4 Benchmarks

- Compiler throughput (pipelines/sec)
- API latency (p50, p99)
- Executor submission latency
- Storage query performance

---

## 11. Future Extensibility

### 11.1 Executor Plugins

- Plugin interface: standard `ExecutorDriver` trait
- Registry: catalog of available executors
- Discovery: load drivers dynamically

### 11.2 Optimization Passes

- Custom optimization rules via plugin interface
- Community-contributed optimizers

### 11.3 Input Formats

- Terraform module → FlowForge pipeline
- dbt → FlowForge lineage
- Custom DSL adapters

---

## 12. Non-Goals (Out of Scope)

- Real-time streaming (batch-focused)
- Machine learning model training (can be tasks, not native)
- Graph databases (PostgreSQL sufficient initially)
- Multi-cloud federation (single cloud provider per deployment)

---

## 13. Glossary

- **IR**: Intermediate Representation (immutable AST)
- **Executor**: Runtime that executes tasks (Argo, Airflow, Local)
- **Driver**: FlowForge adapter for an executor
- **Task**: Unit of work (container image + command)
- **Lineage**: Data provenance graph (task → output, input ← task)
- **Schema Contract**: Data structure agreement between tasks
- **Compiler**: Transforms IR → executor-specific config

