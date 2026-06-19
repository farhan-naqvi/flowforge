# MVP Scope & Phased Roadmap

## 1. MVP (Minimum Viable Product) - Phase 0

**Timeline**: 8 weeks  
**Goal**: Prove the core abstraction (IR → multiple executors)  
**Success Metrics**:
- Single pipeline can execute on both Argo (K8s) and Local dev
- Python SDK usable for basic pipelines
- YAML input format working
- Cost estimation functional

### 1.1 MVP Feature Set

#### Core Engine ✓
- [x] IR specification (Protobuf)
- [x] IR validator (schema, DAG, resource constraints)
- [x] YAML parser → IR
- [x] Python SDK (basic Pipeline API)
- [x] Compiler (AST → Argo/Local config)

#### Execution ✓
- [x] Argo Workflows driver (submit, monitor, logs)
- [x] Local driver (Python subprocess runner)
- [x] Basic status tracking (pending, running, success, failed)

#### API ✓
- [x] gRPC service: Pipeline CRUD
- [x] gRPC service: Execution trigger, status, logs
- [x] REST wrapper (for UI)
- [x] PostgreSQL schema (pipelines, executions)

#### Platform ✓
- [x] Basic authentication (API tokens)
- [x] Error handling + structured logging
- [x] Prometheus metrics (API latency, compiler duration)
- [x] Docker image + docker-compose (local dev)

#### Python SDK ✓
- [x] `Pipeline` class (fluent API)
- [x] `Task` class (image, command, env vars)
- [x] `@flowforge.task` decorator (optional)
- [x] CLI: `ff local pipeline.yaml` (local execution)
- [x] CLI: `ff submit pipeline.yaml --executor argo` (K8s execution)

#### UI (Minimal) ✓
- [x] Dashboard: list pipelines
- [x] Dashboard: execution history
- [x] Dashboard: task logs
- [x] NOT: visual builder (scope reduction)

#### Testing ✓
- [x] Unit tests: compiler, parser, validator (>80% coverage)
- [x] Integration tests: API ↔ DB
- [x] E2E tests: YAML → Argo, SDK → Local
- [x] Fixtures: sample pipelines

#### Documentation ✓
- [x] README (getting started)
- [x] Architecture overview
- [x] SDK reference
- [x] YAML reference
- [x] Deployment guide (K8s + local)

### 1.2 MVP Out-of-Scope

| Feature | Why Deferred | Phase |
|---------|-------------|-------|
| Visual builder (UI drag-drop) | High complexity, not critical for MVP | 1 |
| Airflow driver | Can prove abstraction with Argo + Local | 2 |
| Lineage tracking | Extra complexity, valuable but not MVP | 1 |
| Cost tracking | Post-MVP optimization | 1 |
| Replay/Diff | Requires comprehensive state tracking | 2 |
| Self-healing | Executor-specific feature | 3 |
| Templates | Content feature, not MVP | 2 |
| Multi-tenancy | Auth scope, not MVP | 3 |

### 1.3 MVP Repository Structure (Abbreviated)

```
flowforge/
├── ARCHITECTURE.md
├── api/                        # gRPC + REST server
│   ├── internal/server/
│   ├── proto/                  # IR definitions
│   ├── tests/
│   └── Dockerfile
├── compiler/                   # Parsing, optimization, codegen
│   ├── internal/
│   │   ├── ir/
│   │   ├── parser/
│   │   └── codegen/
│   └── tests/
├── executor/                   # Argo + Local drivers
│   ├── internal/
│   │   ├── argo/
│   │   └── local/
│   └── tests/
├── sdk/                        # Python SDK + CLI
│   ├── flowforge/
│   ├── tests/
│   └── setup.py
├── infra/
│   ├── helm/flowforge/
│   ├── docker-compose.yml
│   └── Dockerfile
├── examples/
│   ├── basic_etl.yaml
│   └── basic_etl.py
├── tests/e2e/
├── docs/
└── Makefile
```

### 1.4 MVP Success Criteria

1. **Functional**:
   - Pipeline defined in Python SDK runs on local machine
   - Same pipeline defined in YAML runs on Argo Workflows
   - Execution history persisted in PostgreSQL
   - Task logs retrievable via API

2. **Performance**:
   - API response time < 100ms (p95)
   - Compiler duration < 500ms (single pipeline)
   - Argo submission latency < 2s

3. **Reliability**:
   - Unit test coverage > 80%
   - E2E tests pass consistently
   - Graceful error handling (no panics)

4. **Documentation**:
   - README has working quick-start
   - Architecture document complete
   - SDK API documented (docstrings)

---

## 2. Phased Roadmap (12 Months)

### Phase 0: MVP (Weeks 1-8)
**Focus**: Core abstraction proof  
**Deliverables**:
- IR specification + compiler
- Argo + Local drivers
- Python SDK + CLI
- Basic API + Dashboard
- Docker deployment

**Team**: 2-3 engineers (backend), 1 frontend

---

### Phase 1: Enhanced Capabilities (Weeks 9-16)
**Focus**: User-facing features  

**Features**:

1. **Visual Builder** (UI)
   - Drag-drop task canvas
   - Property editor (image, command, env, resources)
   - Connection drawing (DAG edges)
   - Save → YAML
   - Components: Canvas, TaskNode, Toolbar

2. **Lineage & Debugging**
   - Execution graph visualization
   - Task-to-task data flow
   - Input/output inspection
   - Provenance query API
   - Lineage store (PostgreSQL)

3. **Replay & Comparison**
   - Re-run execution with same inputs
   - Diff previous runs (IR changes, cost, duration)
   - Partial replay (from task X onward)

4. **Python SDK Enhancements**
   - Type hints + Pydantic validation
   - `@flowforge.task` decorator (implicit Task creation)
   - Schema contracts (JSON Schema / Pydantic)
   - Conditional execution (if/else branches)
   - Parametrization (templated inputs)

5. **Cost Tracking**
   - Per-task resource allocation
   - Cost estimation (before execution)
   - Cost tracking (actual usage)
   - Dashboard: cost breakdown by task/executor
   - Cost alerts

6. **Local Mode Enhancements**
   - Watch mode (auto-rerun on file changes)
   - Profiling (task duration, memory)
   - Failure simulation

**Deliverables**:
- Fully functional visual builder
- Lineage query API + UI explorer
- Replay/diff functionality
- Cost dashboard
- Enhanced SDK

**Team**: 3 engineers (2 backend, 1 frontend)

---

### Phase 2: Executor Diversity (Weeks 17-24)
**Focus**: Multi-executor support  

**Features**:

1. **Airflow Driver**
   - DAG generation from IR
   - XCom integration (data passing)
   - Task dependencies mapping
   - Scheduling policies (Airflow sensors, SLAs)
   - Monitoring via Airflow UI
   - Status sync back to FlowForge

2. **Executor Benchmarking**
   - Run same pipeline on Argo vs Airflow
   - Compare: latency, cost, resource usage
   - Metrics: overhead, queuing delay
   - Recommendation engine (which executor for this workload)

3. **Advanced Scheduling**
   - Trigger policies (on-schedule, on-event, manual)
   - SLA monitoring (task duration exceeds threshold)
   - Backfill (historical date range execution)
   - Dynamic task generation (scatter-gather)

4. **Data Contract Enforcement**
   - Schema validation at task boundaries (runtime)
   - Schema change detection (breaking changes)
   - Schema versioning + documentation
   - Data quality rules (nullability, value ranges)

**Deliverables**:
- Production-ready Airflow driver
- Benchmarking suite
- Advanced scheduling capabilities
- Schema contract system

**Team**: 3 engineers (2 backend, 1 infrastructure)

---

### Phase 3: Enterprise Features (Weeks 25-36)
**Focus**: Production hardening  

**Features**:

1. **Self-Healing**
   - Automatic retry with backoff
   - Fallback tasks (alternative implementations)
   - Partial failure handling (skip failures, continue)
   - Anomaly detection (task duration > 3σ)
   - Auto-remediation (restart, scale resources)

2. **Advanced Lineage**
   - Cross-pipeline lineage (pipeline A output → pipeline B input)
   - Schema propagation (infer schema through lineage)
   - Data quality metrics (nulls, duplicates)
   - Cost attribution (by data source)
   - Compliance tracking (PII data, retention)

3. **Integration Ecosystem**
   - Airflow operator for FlowForge pipelines (call as external DAG)
   - dbt adapter (convert dbt DAG → FlowForge)
   - Terraform provider (manage pipelines as IaC)
   - Kubernetes CRD (CRD for pipelines)
   - Webhook triggers (external event → pipeline)

4. **Multi-Tenancy & Security**
   - Organization isolation
   - RBAC (viewer, editor, admin, owner per org)
   - Audit logging (who did what, when)
   - Secret management (Vault integration)
   - Network policies (pod-to-pod, egress control)

5. **Advanced UI**
   - Pipeline versioning UI
   - Diff viewer (before/after pipeline)
   - Resource usage breakdown (CPU, memory, storage)
   - Real-time execution streaming
   - Collaborative editing (multiple users)

**Deliverables**:
- Self-healing engine
- Cross-pipeline lineage
- Integration adapters
- Enterprise auth
- Advanced UI

**Team**: 4-5 engineers (3 backend, 1-2 frontend, 1 infrastructure)

---

### Phase 4: Scaling & Optimization (Weeks 37-48)
**Focus**: Performance, cost, scale  

**Features**:

1. **Performance Optimization**
   - Compiler caching (memoize parse/optimize)
   - IR compression (reduce storage)
   - Executor-specific optimizations
     - Argo: pod optimization, image layer caching
     - Airflow: DAG serialization caching
   - Database query optimization (indexes, materialized views)

2. **Cost Optimization**
   - Spot instance support (Kubernetes Spot)
   - Preemptible task detection
   - Resource right-sizing recommendations
   - Multi-zone cost analysis
   - Reserved capacity planning

3. **Advanced Runtimes**
   - Ray (distributed Python tasks)
   - Spark (big data transformations)
   - DuckDB (in-process analytics)
   - FaaS (AWS Lambda, Google Cloud Functions)
   - Container-less executor (WASM)

4. **Observability Enhancements**
   - Distributed tracing (full request path)
   - Custom metrics (business KPIs)
   - Anomaly detection (statistical baselines)
   - Root cause analysis (log correlation)
   - SLO monitoring (uptime, latency SLOs)

5. **Developer Experience**
   - VSCode extension (lint pipelines, syntax highlight)
   - IDE integration (autocomplete, type hints)
   - Debugging tools (breakpoints, step-through)
   - Local hot-reload (ff watch)
   - Version management (semantic versioning)

**Deliverables**:
- Multi-runtime support
- Cost optimization engine
- Advanced observability
- Developer tooling

**Team**: 3-4 engineers (focus on backend optimization, DevX)

---

## 3. Roadmap Timeline

```
Phase 0 (MVP)          Weeks 1-8
├─ IR + Compiler       ✓
├─ Argo + Local        ✓
├─ Python SDK          ✓
├─ API + Dashboard     ✓
└─ Docker deployment   ✓

Phase 1 (UX)           Weeks 9-16
├─ Visual Builder      ✓
├─ Lineage             ✓
├─ Replay/Diff         ✓
├─ Schema Contracts    ✓
└─ Cost Tracking       ✓

Phase 2 (Multi-Exec)   Weeks 17-24
├─ Airflow Driver      ✓
├─ Benchmarking        ✓
├─ Advanced Scheduling ✓
└─ Data Contracts      ✓

Phase 3 (Enterprise)   Weeks 25-36
├─ Self-Healing        ✓
├─ Cross-Pipeline      ✓
├─ Integrations        ✓
├─ Multi-Tenancy       ✓
└─ Advanced UI         ✓

Phase 4 (Scale)        Weeks 37-48
├─ Performance Opt     ✓
├─ Cost Opt            ✓
├─ Ray/Spark/FaaS      ✓
├─ Observability       ✓
└─ Developer Tools     ✓
```

---

## 4. Dependency Relationships

```
Phase 0 (Foundation)
    ↓
Phase 1 (UX Features, depends on Phase 0)
    ├─ Visual Builder (requires IR stable)
    ├─ Lineage (requires execution history)
    └─ Cost Tracking (requires metrics infrastructure)
    ↓
Phase 2 (Multi-Executor, depends on Phase 0+1)
    ├─ Airflow Driver (requires IR stable, codegen pattern)
    ├─ Benchmarking (requires execution metrics)
    └─ Advanced Scheduling (requires executor flexibility)
    ↓
Phase 3 (Enterprise, depends on Phase 0+1+2)
    ├─ Self-Healing (requires monitoring + retry logic)
    ├─ Integrations (requires stable APIs)
    └─ Multi-Tenancy (requires auth framework)
    ↓
Phase 4 (Scale, depends on Phase 0-3)
    ├─ Runtime Plugins (requires driver abstraction)
    └─ DevX (requires mature platform)
```

---

## 5. Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|-----------|
| IR design mistakes | High | Mock multiple executors early; validate IR expressiveness |
| Argo learning curve | Medium | Dedicated Argo expert on team; spike on K8s operators |
| Performance (compiler) | Medium | Benchmark early; profiling in CI |
| Airflow incompatibility | High | Start Phase 2 early; map IR → Airflow DAG concepts |
| PostgreSQL scaling | Medium | Partition execution history; consider time-series DB later |
| Team hiring | High | Start recruiting phase 1; focus on Go expertise |

---

## 6. Success Metrics by Phase

### Phase 0
- [ ] Compiler throughput: 100+ pipelines/sec
- [ ] API p95 latency < 100ms
- [ ] Unit test coverage > 80%
- [ ] E2E test pass rate 100%
- [ ] Documented quick-start (< 10 min from clone to first execution)

### Phase 1
- [ ] Builder used for 50%+ of new pipelines (vs SDK)
- [ ] Lineage API queries < 500ms (p95)
- [ ] Replay success rate > 95%
- [ ] Cost estimates within ±20% of actual

### Phase 2
- [ ] Airflow driver parity with Argo (feature coverage)
- [ ] Benchmarking data available for top 10 workload types
- [ ] Adoption: 30%+ of pipelines use Airflow executor

### Phase 3
- [ ] Self-healing reduces manual intervention by 70%
- [ ] Multi-tenancy supports 100+ orgs
- [ ] Airflow integration: > 5 active users

### Phase 4
- [ ] Ray/Spark tasks: 20%+ of advanced workloads
- [ ] Cost optimization: 30% average cost reduction
- [ ] Developer tools: 80%+ adoption (VSCode extension)

---

## 7. Resource Planning

### Phase 0 (MVP)
- **Backend**: 2-3 engineers (Go)
- **Frontend**: 1 engineer (React) - minimal UI
- **DevOps**: 0.5 engineer (Docker, docker-compose)
- **Total**: 3.5 FTE

### Phase 1 (UX)
- **Backend**: +1 engineer
- **Frontend**: +1 engineer (visual builder)
- **DevOps**: +0.5 engineer (observability)
- **Total**: 6 FTE

### Phase 2 (Multi-Executor)
- **Backend**: +1 engineer (Airflow)
- **Platform**: +1 engineer (benchmarking)
- **Total**: 8 FTE

### Phase 3 (Enterprise)
- **Backend**: +1 engineer (self-healing, integrations)
- **Platform**: +1 engineer (security, multi-tenancy)
- **Total**: 10 FTE

### Phase 4 (Scale)
- **Backend**: +1 engineer (optimization)
- **DevX**: +1 engineer (developer tools)
- **Total**: 12 FTE

---

## 8. Budget Estimation (Rough)

| Category | Phase 0 | Phase 1 | Phase 2 | Phase 3 | Phase 4 |
|----------|---------|---------|---------|---------|---------|
| Salaries (FTE × $120k/yr) | $420k | $720k | $960k | $1.2M | $1.44M |
| Infrastructure (K8s, DB, etc) | $50k | $100k | $150k | $200k | $250k |
| Tools & Services (monitoring, CI/CD) | $20k | $40k | $60k | $80k | $100k |
| **Phase Total** | **$490k** | **$860k** | **$1.17M** | **$1.48M** | **$1.79M** |
| **Cumulative** | **$490k** | **$1.35M** | **$2.52M** | **$4M** | **$5.79M** |

*Note: Rough estimates; assumes $120k average TC, US-based costs*

