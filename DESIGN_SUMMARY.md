# FlowForge Design Summary & Key Decisions

## Executive Summary

FlowForge is a **declarative data pipeline orchestration platform** that abstracts away executor specifics through an Intermediate Representation (IR). The design enables users to define pipelines once (Python SDK, YAML, or UI) and execute on multiple runtimes (Argo Workflows, Airflow, local dev).

**Core Innovation**: Single IR → multiple executors (Argo, Airflow, Local) without rewriting pipelines.

---

## Key Architectural Decisions

### 1. Intermediate Representation (IR) as Foundation
✓ **Decision**: Define IR as immutable, versioned AST with task/DAG/schema/lineage metadata  
**Why**: Language-agnostic, enables multi-executor support, auditable for compliance  
**Trade-off**: Requires upfront domain modeling vs. DSL-only approach  
**Alternative Rejected**: Direct YAML → Executor compilation (loses abstraction)

### 2. Multi-Stage Compiler
✓ **Decision**: Parse → Validate → Optimize → Codegen pipeline  
**Why**: Separation of concerns, pluggable rules, executor-specific optimization  
**Trade-off**: Complexity upfront vs. future flexibility  
**Alternative Rejected**: Monolithic single-pass compiler

### 3. Go Backend, Python SDK
✓ **Decision**: Control plane in Go (gRPC/REST), Python SDK wraps it, transformations in containers  
**Why**: Go handles concurrency/gRPC natively, Python is data engineer friendly  
**Trade-off**: Polyglot complexity vs. language homogeneity  
**Alternative Rejected**: Pure Python backend (performance concerns at scale)

### 4. Three Execution Modes
✓ **Decision**: Local (dev), Argo (cloud-native), Airflow (enterprise)  
**Why**: Dev velocity, cloud-native serverless, enterprise integration  
**Trade-off**: Must maintain driver abstractions  
**Alternative Rejected**: Single executor (less flexible)

### 5. PostgreSQL + Redis
✓ **Decision**: PostgreSQL for state/lineage/history, Redis for caching/queuing  
**Why**: ACID compliance, JSON support, mature ecosystem  
**Trade-off**: Relational model less flexible than document DBs  
**Alternative Rejected**: MongoDB (schema structure benefits SQL)

### 6. Executor Driver Pattern
✓ **Decision**: Define `ExecutorDriver` interface, implement per-executor  
**Why**: Enables new executors without core changes (Ray, Spark, FaaS)  
**Trade-off**: Requires contract enforcement, versioning complexity  
**Alternative Rejected**: Direct Argo/Airflow integration in core

### 7. Not an Airflow Replacement
✓ **Decision**: Complement Airflow, not replace it  
**Why**: Respect existing investments, integrate where it makes sense  
**Trade-off**: More complex feature parity requirements  
**Alternative Rejected**: Pure Airflow replacement (alienates existing users)

---

## System Layers

```
┌─────────────────────────────────────────┐
│ UI Layer (React/TypeScript)             │
│ Builder | Explorer | Dashboard          │
└─────────────────────────────────────────┘
                    ↓ REST/gRPC
┌─────────────────────────────────────────┐
│ API Layer (Go)                          │
│ Pipeline | Execution | Lineage | Cost   │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│ Core Engine (Go)                        │
│ Compiler | Executor Drivers | Lineage   │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│ Runtime Layer                           │
│ Argo Workflows | Airflow | Local        │
└─────────────────────────────────────────┘
```

---

## Critical Abstractions

### Parser Interface
Multiple input formats → IR (YAML, Python SDK, UI events)

### ExecutorDriver Interface
IR → executor submission, monitoring, logging

### OptimizationPass Interface
Pluggable IR transformations

### CodeGenerator Interface
IR → executor-specific config (Argo YAML, Airflow DAG, Local DAG)

---

## MVP (8 Weeks) Features

**In Scope**:
- IR + Compiler (Argo + Local codegen)
- Python SDK + CLI
- Argo Workflows driver
- Local dev driver
- Basic API + Dashboard
- PostgreSQL persistence
- Docker deployment

**Out of Scope** (Phase 1+):
- Visual builder (complex)
- Airflow driver (proves abstraction after MVP)
- Lineage tracking (post-MVP)
- Replay/Diff (depends on execution state)
- Self-healing (executor-specific)

---

## Repository Structure Highlights

```
flowforge/
├── ARCHITECTURE.md                 (This design)
├── api/                            (gRPC + REST server)
├── compiler/                       (Parse, validate, compile)
├── executor/                       (Argo + Local drivers)
├── sdk/                            (Python SDK + CLI)
├── ui/                             (React dashboard, builder)
├── infra/                          (Terraform + Helm)
├── examples/                       (Sample pipelines)
├── tests/                          (e2e tests)
└── docs/                           (User & dev guides)
```

**Key Principle**: Each module independently deployable via semver.

---

## Design Patterns Used

### 1. Dependency Injection
Compiler receives drivers, parsers, generators as interfaces (not concrete types).

```go
compiler := NewCompiler(
    parser,          // Parser interface
    validators,      // []ValidationRule
    optimizers,      // []OptimizationPass
    generators,      // map[string]CodeGenerator
)
```

### 2. Builder Pattern
Python SDK uses fluent builder:

```python
pipeline = (Pipeline("etl")
    .add_task("extract", image="postgres-client")
    .add_task("transform", image="python:3.11")
    .connect("extract", "transform")
    .submit())
```

### 3. Strategy Pattern
Different executors (Argo, Local, Airflow) implement `ExecutorDriver` interface.

### 4. Factory Pattern
CodeGen factory creates executor-specific generators.

### 5. Pipeline Pattern
Multi-stage compiler: Parse → Validate → Optimize → Codegen

---

## Observability Plan

**Metrics** (Prometheus):
- API latency (gRPC/REST)
- Compiler duration (parse, optimize, codegen)
- Execution metrics (task success rate, duration)
- Storage metrics (query latency, connection pool)

**Logging** (Structured JSON):
- API logs (request/response)
- Compiler logs (optimization decisions)
- Execution logs (stderr/stdout from tasks)

**Tracing** (OpenTelemetry):
- Full request path: API → Compiler → Executor
- Task execution trace
- Database query tracing

**Dashboards** (Grafana):
- System health (API, DB, Executor uptime)
- Pipeline performance (p50, p95 latencies)
- Cost tracking (per task, per executor)
- Error rates + alerts

---

## Security Model

**Authentication**:
- OIDC (GitHub, Google, enterprise SSO)
- API tokens for programmatic access

**Authorization**:
- RBAC (viewer, editor, admin per organization/pipeline)
- Audit logging (who did what, when)

**Data Protection**:
- TLS for all network communication
- PostgreSQL encryption at rest
- Secret management (Vault integration)

**Runtime Security**:
- Container image signing (Cosign)
- Network policies (K8s NetworkPolicy)
- Resource limits (CPU, memory, timeout)

---

## Extensibility Points

### 1. New Executors
Implement `ExecutorDriver` interface:
```go
type ExecutorDriver interface {
    Submit(ctx context.Context, config *ExecutorConfig) (*Submission, error)
    Status(ctx context.Context, submissionID string) (*ExecutionStatus, error)
    Logs(ctx context.Context, submissionID, taskID string) (io.Reader, error)
    Cancel(ctx context.Context, submissionID string) error
}
```

### 2. New Input Formats
Implement `Parser` interface:
```go
type Parser interface {
    Parse(ctx context.Context, input interface{}) (*ir.PipelineSpec, error)
    Supports(format string) bool
}
```

### 3. Optimization Rules
Implement `OptimizationPass`:
```go
type OptimizationPass interface {
    Optimize(ctx context.Context, pipeline *ir.PipelineSpec) (*ir.PipelineSpec, error)
    Name() string
    AppliesTo(executor string) bool
}
```

### 4. Custom Metrics/Observability
Inject observability collectors into core components.

---

## Testing Strategy

**Unit Tests** (80%+ coverage):
- IR validation rules
- Parser correctness
- Optimizer behavior
- Cost calculations

**Integration Tests** (module pairs):
- Parser → Compiler
- Compiler → Driver
- Driver → Executor
- API → Database

**E2E Tests** (full flows):
- YAML → Argo Workflow → K8s execution
- SDK → Local execution
- UI builder → Pipeline execution
- Execution → Lineage storage

**Performance Tests**:
- Compiler throughput (100+ pipelines/sec target)
- API latency (p95 < 100ms target)
- Storage query performance

---

## Phased Roadmap

| Phase | Timeline | Focus | Deliverables |
|-------|----------|-------|--------------|
| **0 (MVP)** | Weeks 1-8 | Core abstraction | IR, Compiler, Argo+Local drivers, Python SDK |
| **1 (UX)** | Weeks 9-16 | User features | Visual builder, lineage, replay, cost tracking |
| **2 (Multi-Exec)** | Weeks 17-24 | Executor diversity | Airflow driver, benchmarking, scheduling |
| **3 (Enterprise)** | Weeks 25-36 | Production features | Self-healing, integrations, multi-tenancy |
| **4 (Scale)** | Weeks 37-48 | Optimization | Ray/Spark/FaaS, cost optimization, DevX |

---

## Design Tradeoffs Summary

| Aspect | Choice | Benefit | Cost |
|--------|--------|---------|------|
| **Language** | Go backend | Concurrency, gRPC native, performance | Learning curve |
| **Serialization** | Protobuf | Versioning, cross-language | Binary format |
| **Storage** | PostgreSQL | ACID, JSON, mature | Schema rigidity |
| **Primary Executor** | Argo | K8s native, serverless | K8s dependency |
| **Python Runtime** | Containerized | Isolation, reproducibility | Container overhead |
| **Input Formats** | Multi-format | Flexibility, low barrier | Parser complexity |
| **Execution Models** | Multiple | Developer + enterprise flexibility | Driver abstraction complexity |

**Rationale**: Each choice optimizes for extensibility and multi-executor support while maintaining production-ready performance.

---

## Next Steps (If Approved)

1. **Week 1**: Team kickoff, development environment setup
2. **Week 1-2**: IR specification finalization (Proto files)
3. **Week 2-3**: Core compiler module (parser, validator, codegen stubs)
4. **Week 3-4**: Argo driver implementation
5. **Week 4-5**: Local driver + Python SDK basic
6. **Week 5-6**: API server + database schema
7. **Week 6-7**: Basic UI dashboard
8. **Week 8**: Integration testing, hardening, release

---

## Design Review Checklist

- [x] IR specification complete (covers all pipeline types)
- [x] Executor abstraction viable (Argo + Local + Airflow mappable)
- [x] API design (gRPC services, REST wrappers)
- [x] Database schema (pipelines, executions, lineage)
- [x] Security model (auth, encryption, secrets)
- [x] Observability (metrics, logs, traces)
- [x] Testing strategy (unit, integration, e2e)
- [x] Documentation (architecture, SDK, deployment)
- [x] Scalability analysis (concurrency, storage)
- [x] Extensibility (plugins, custom executors)
- [x] MVP scope (achievable in 8 weeks)
- [x] Roadmap alignment (phases, dependencies, resource planning)

---

## Questions for Team Discussion

1. **IR Expressiveness**: Are task-level scheduling policies, resource sharing patterns, and data contracts sufficient for your use cases?

2. **Executor Priority**: Should we build Airflow driver in Phase 2, or is Argo + Local MVP sufficient?

3. **UI Scope**: Should visual builder be Phase 1 feature, or is YAML+SDK sufficient for MVP?

4. **Data Contracts**: Do we want to enforce schema contracts at runtime (performance cost) or just validate?

5. **Cost Model**: Should cost tracking be per-task, per-pipeline, or both?

6. **Multi-Tenancy**: Is org-level isolation needed, or start with single-tenant MVP?

7. **Lineage Depth**: Should we track data lineage across multiple runs, or just within single execution?

---

## Conclusion

This design proposes **FlowForge** as a **portfolio-grade data pipeline platform** with these defining characteristics:

✓ **Abstraction layer** (IR) decouples pipeline definition from execution  
✓ **Multi-executor support** (start with Argo + Local, add Airflow later)  
✓ **Developer-friendly** (Python SDK, YAML, visual UI)  
✓ **Enterprise-ready** (lineage, replay, cost visibility, security)  
✓ **Extensible** (plugin architecture for new executors, optimizations)  
✓ **Achievable MVP** (8 weeks, 3-4 engineers)

The design balances **immediate viability** (MVP in 8 weeks) with **long-term ambition** (multi-executor, enterprise features, optimal performance).

