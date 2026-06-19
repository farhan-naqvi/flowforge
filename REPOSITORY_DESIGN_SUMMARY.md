# FlowForge Repository Design - Summary

## Overview

FlowForge is organized as a **monorepo** with 11 core Go/Python modules, each with clear ownership, interfaces, and dependencies. This document summarizes the key design principles and structure.

---

## 🏗️ Monorepo Modules (11 Core)

### Foundational Modules

1. **ir/** — Intermediate Representation
   - Core IR spec, validation, serialization
   - Go module (standalone)
   - Owner: Core Team
   - Exports: `PipelineSpec`, `TaskSpec`, `Validator`, `IRBuilder`

2. **compiler/** — Compilation Pipeline
   - Parse (YAML/SDK/UI → IR)
   - Validate (spec, DAG, resources)
   - Optimize (task merging, parallelism)
   - Codegen (IR → Argo/Airflow/Local)
   - Go module (depends: ir/)
   - Owner: Compiler Team
   - Exports: `Parser`, `Compiler`, `OptimizationPass`, `CodeGenerator`

3. **runtime/** — Local Execution
   - Task runner (subprocess execution)
   - DAG dependency resolver
   - Result collection & caching
   - Go module (depends: ir/, compiler/)
   - Owner: Execution Team
   - Exports: `Runner`, `Executor`, `TaskExecutor`

### Execution Modules

4. **executors/** — Executor Drivers
   - Argo Workflows driver
   - Airflow driver
   - Local driver (wraps runtime/)
   - Go module (depends: ir/, compiler/)
   - Owner: Execution Team
   - Exports: `ExecutorDriver`, `SubmissionRequest`, `ExecutionStatus`

### Data & Persistence

5. **storage/** — Persistence Layer
   - PostgreSQL client (pipelines, executions, metrics)
   - Redis client (caching, queuing)
   - Schema migrations
   - Query builders
   - Go module (standalone)
   - Owner: Data Team
   - Exports: `Store`, `Query`, `Transaction`

6. **lineage/** — Provenance Tracking
   - Lineage graph construction
   - Data flow tracking
   - PostgreSQL lineage storage
   - Graph queries
   - Go module (depends: ir/, storage/)
   - Owner: Data Team
   - Exports: `LineageEngine`, `Graph`, `Node`, `Edge`

### Platform Services

7. **observability/** — Metrics, Logs, Traces
   - Prometheus metrics registry
   - Structured JSON logging
   - OpenTelemetry tracing
   - Health checks
   - Go module (standalone, injected via middleware)
   - Owner: Platform Team
   - Exports: `MetricsCollector`, `Logger`, `Tracer`, `HealthChecker`

8. **api/** — API Server
   - gRPC server (Protocol Buffers)
   - REST gateway (gRPC-Gateway)
   - Request handlers
   - Business logic services
   - Middleware (auth, RBAC, logging, metrics)
   - Go module (depends: all core modules)
   - Owner: Platform Team
   - Exports: gRPC services, REST routes

### Client-Side & Infrastructure

9. **sdk/** — Python SDK + CLI
   - High-level Pipeline builder API
   - Decorators (@flowforge.task)
   - CLI commands (ff submit, ff local, etc)
   - gRPC/REST client wrapper
   - Python package (depends: api/ via gRPC)
   - Owner: SDK Team
   - Exports: `Pipeline`, `Task`, `Client`, `decorators`

10. **ui/** — Frontend (React + TypeScript)
    - Visual pipeline builder (canvas, nodes)
    - Execution explorer (lineage, replay)
    - Dashboard (pipelines, executions, costs)
    - REST client integration
    - React SPA (depends: api/ via REST)
    - Owner: UI Team
    - Exports: React components, custom hooks

11. **deployment/** — Infrastructure as Code
    - Terraform modules (K8s, DB, networking)
    - Helm charts (K8s manifests)
    - Docker images (API, Executor)
    - docker-compose (local dev)
    - CI/CD scripts
    - Owner: DevOps Team
    - Exports: Infrastructure definitions

### Supporting Modules

12. **examples/** — Sample Pipelines
    - ETL workflow
    - Data quality pipeline
    - ML workflow
    - Multi-executor examples

13. **tests/** — E2E Integration Tests
    - Cross-module tests
    - Test fixtures & factories
    - Test data

14. **docs/** — Documentation
    - User guides
    - Development guides
    - API reference
    - Deployment guides

---

## 🔄 Dependency Graph

```
Minimal & Acyclic:

                SDK (Python)
                     │ gRPC
                 API Server
                /    │    │    \
            Compiler │  Storage  │
            /   │    │     │      │
        Parser  │    │     │   Lineage
             IR-┘    │     │
              └──────┴─────┴─ Observability
                    │
                Executors
                 /  │  \
              Argo AF Local
                 │
              Runtime

No circular dependencies (acyclic DAG).
Observability is cross-cutting (injected).
Storage is standalone (all modules use it).
```

---

## 📦 Directory Structure

```
flowforge/
├── ir/              [Go] Core IR spec, validation
├── compiler/        [Go] Parse, compile, optimize, codegen
├── runtime/         [Go] Local task execution
├── executors/       [Go] Argo, Airflow, Local drivers
├── storage/         [Go] PostgreSQL, Redis persistence
├── lineage/         [Go] Data provenance tracking
├── observability/   [Go] Metrics, logs, traces
├── api/             [Go] gRPC + REST server
├── sdk/             [Python] SDK + CLI
├── ui/              [React/TS] Dashboard & builder
├── deployment/      [HCL/YAML] Terraform, Helm, Docker
├── examples/        [YAML/Python] Sample pipelines
├── tests/           [Python/Go] E2E integration tests
├── docs/            [Markdown] User & dev guides
├── scripts/         [Bash] Development scripts
├── .github/         GitHub Actions, issue templates
├── go.work          Go workspace (monorepo)
├── Makefile         Root-level tasks
├── docker-compose.yml Local dev environment
├── README.md        Monorepo overview
├── ARCHITECTURE.md  System design
├── MODULE_BOUNDARIES.md Module ownership & contracts
├── CONTRIBUTING.md  Contribution guidelines
└── VERSION          Version tag
```

---

## 🎯 Key Design Principles

### 1. Module Independence
- Each module deployable independently
- Clear public API (pkg/)
- Private implementation (internal/)
- Semantic versioning

### 2. Acyclic Dependencies
- ir/ has no dependencies (foundation)
- compiler/ depends on ir/
- executors/ depend on ir/ + compiler/
- api/ depends on all (orchestrator)
- No circular dependencies allowed

### 3. Interface-Based Design
- Depend on interfaces, not implementations
- All module communication via public API
- Pluggable components (parsers, optimizers, drivers)
- Registry pattern for extensions

### 4. Separation of Concerns
- ir/: specification
- compiler/: transformation
- executors/: execution
- storage/: persistence
- observability/: cross-cutting
- api/: orchestration

### 5. Minimal Coupling
- Modules import from pkg/ only
- No internal/ package imports across modules
- Registry pattern for plugin discovery
- Interface contracts enforce boundaries

---

## 🔧 Extensibility Points

### Add New Executor

1. Implement `ExecutorDriver` interface (executors/pkg/)
2. Register in driver registry
3. Add tests & documentation

### Add New Parser Format

1. Implement `Parser` interface (compiler/pkg/)
2. Register in parser registry
3. Add tests & documentation

### Add Optimization Pass

1. Implement `OptimizationPass` interface (compiler/pkg/)
2. Register in optimizer
3. Add tests & documentation

### Add Observability Collector

1. Implement collector interface (observability/pkg/)
2. Inject into modules
3. Add tests & documentation

---

## 📊 Module Ownership Matrix

| Module | Owner | Team Size | Tech | Exports |
|--------|-------|-----------|------|---------|
| ir/ | Core | 1 | Go | PipelineSpec, Validator |
| compiler/ | Compiler | 2 | Go | Parser, Compiler, CodeGenerator |
| runtime/ | Execution | 1 | Go | Runner, Executor |
| executors/ | Execution | 2 | Go | ExecutorDriver (Argo, Airflow) |
| storage/ | Data | 1 | Go | Store, Query |
| lineage/ | Data | 1 | Go | LineageEngine, Graph |
| observability/ | Platform | 1 | Go | Metrics, Logger, Tracer |
| api/ | Platform | 2 | Go | gRPC services, REST routes |
| sdk/ | SDK | 1 | Python | Pipeline, Task, Client |
| ui/ | UI | 2 | React/TS | Components, hooks, store |
| deployment/ | DevOps | 1 | HCL/YAML | Infrastructure |

---

## 🛠️ Getting Started

### 1. Setup
```bash
make setup-dev
```

### 2. Develop
```bash
cd <module>
go test ./...       # Go modules
pytest tests/ -v    # Python modules
npm test            # React module
```

### 3. Test All
```bash
make test           # All tests
make test-coverage  # Coverage report
```

### 4. Submit
```bash
git checkout -b feature/name
# Make changes + tests
make lint && make fmt
git push
# Create PR (see CONTRIBUTING.md)
```

---

## 📚 Key Files

| File | Purpose |
|------|---------|
| [README.md](README.md) | Monorepo overview, quick start |
| [ARCHITECTURE.md](ARCHITECTURE.md) | System design, decisions, tradeoffs |
| [MODULE_BOUNDARIES.md](MODULE_BOUNDARIES.md) | Module ownership, interfaces, contracts |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute, code standards |
| [Makefile](Makefile) | Build, test, lint, deploy commands |
| [go.work](go.work) | Go monorepo configuration |
| [docker-compose.yml](docker-compose.yml) | Local dev environment |

---

## ✅ Validation Checklist (Before Implementation)

- [x] 11 modules identified with clear ownership
- [x] Acyclic dependencies enforced
- [x] Public API defined (interfaces in pkg/)
- [x] Private implementation (internal/)
- [x] Extensibility points identified
- [x] Monorepo structure designed
- [x] Go workspace (go.work) configured
- [x] Makefile for cross-module tasks
- [x] docker-compose for local dev
- [x] Contributing guidelines documented
- [x] Module boundaries documented
- [x] No circular dependencies

---

## 🚀 Next Steps (Implementation)

### Week 1: Foundation
1. Create directory structure (`make dirs`)
2. Initialize Go modules (go.mod per module)
3. Define public interfaces (pkg/)
4. Setup testing infrastructure

### Week 2-3: Core Engine
1. IR specification (ir/)
2. YAML parser (compiler/)
3. Argo codegen (compiler/)
4. Argo driver (executors/)

### Week 4-5: API & SDK
1. API server (api/)
2. PostgreSQL schema (storage/)
3. Python SDK (sdk/)

### Week 6-7: Integration
1. E2E tests (tests/)
2. Docker images (deployment/)

### Week 8: Release
1. Documentation
2. MVP release (v0.1.0)

---

## 📝 Summary

FlowForge's repository design balances **flexibility** with **structure**:

✓ **Clear ownership**: Each module has assigned team  
✓ **Acyclic dependencies**: No circular dependencies  
✓ **Interface-based**: Pluggable components  
✓ **Independent deployability**: Modules can ship separately  
✓ **Extensible**: New executors, parsers, optimizers via interfaces  
✓ **Minimal coupling**: Public API isolation  
✓ **Well-documented**: README, guides, contracts  

This design enables **parallel development**, **easy testing**, and **future extensibility** (Ray, Spark, FaaS executors).

