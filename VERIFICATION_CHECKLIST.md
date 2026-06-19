# FlowForge Monorepo Verification Checklist

Use this checklist to verify the repository structure and design principles are followed during implementation.

---

## 📋 Pre-Implementation

- [ ] All 11 modules planned (ir, compiler, runtime, executors, storage, lineage, observability, api, sdk, ui, deployment)
- [ ] Each module has assigned owner (see MODULE_BOUNDARIES.md)
- [ ] Module ownership documented in team wiki/wiki
- [ ] Communication plan for cross-module changes (weekly syncs)

---

## 🏗️ Directory Structure (Post-Creation)

### Root Level
- [ ] README.md (monorepo overview)
- [ ] ARCHITECTURE.md (system design)
- [ ] MODULE_BOUNDARIES.md (ownership & contracts)
- [ ] REPOSITORY_STRUCTURE.md (directory tree)
- [ ] CONTRIBUTING.md (contribution guidelines)
- [ ] LICENSE (Apache 2.0)
- [ ] .gitignore (covers Go, Python, Node, Terraform)
- [ ] Makefile (cross-module tasks)
- [ ] go.work (Go workspace)
- [ ] docker-compose.yml (local dev environment)
- [ ] VERSION file (0.1.0-alpha)

### Modules (Each module should have)
- [ ] README.md (module purpose, usage)
- [ ] Makefile (module-specific tasks)
- [ ] go.mod + go.sum (Go modules only)
- [ ] setup.py + requirements.txt (Python modules only)
- [ ] package.json (Node modules only)
- [ ] internal/ (private implementation)
- [ ] pkg/ (public API)
- [ ] tests/ (unit, integration, fixtures)

### Core Go Modules
- [ ] ir/
- [ ] compiler/
- [ ] runtime/
- [ ] executors/
- [ ] storage/
- [ ] lineage/
- [ ] observability/
- [ ] api/

### Python/JavaScript Modules
- [ ] sdk/
- [ ] ui/

### Infrastructure & Docs
- [ ] deployment/ (Terraform, Helm, Docker)
- [ ] examples/ (sample pipelines)
- [ ] tests/ (E2E integration tests)
- [ ] docs/ (user & dev guides)
- [ ] scripts/ (dev scripts)
- [ ] .github/ (GitHub Actions, templates)

---

## 🔐 Dependency Rules

### No Circular Dependencies
- [ ] ir/ imported by: compiler, runtime, executors, lineage, storage (expected)
- [ ] ir/ imports from: (none)
- [ ] compiler/ imported by: api, sdk (expected)
- [ ] compiler/ imports from: ir (only)
- [ ] executors/ imported by: api (expected)
- [ ] executors/ imports from: ir, compiler (only)
- [ ] storage/ imported by: all (expected, independent)
- [ ] storage/ imports from: (none except internal)
- [ ] api/ imported by: sdk, ui (expected)
- [ ] api/ imports from: all core modules (expected)
- [ ] Run `go mod graph` to visualize
- [ ] No cycles detected by `go mod graph`

### No Internal Package Imports Across Modules
- [ ] No imports of `module/internal/*` in other modules
- [ ] All cross-module imports use `module/pkg/*`
- [ ] Linter rule: forbid imports of internal/
- [ ] Verify with: `grep -r "internal/" --include="*.go" | grep import`

### Public API Only
- [ ] Each module's pkg/ defines interfaces
- [ ] Only interfaces exported, not implementations
- [ ] internal/ contains implementations
- [ ] public interfaces in pkg/
- [ ] Example: `type Parser interface` in compiler/pkg/compiler.go

---

## 🎯 Module Interface Contracts

### ir/ Module
- [ ] Exports: `PipelineSpec`, `TaskSpec`, `Edge`, `ValidationError`
- [ ] Exports: `Validator` interface
- [ ] Exports: `IRBuilder` interface
- [ ] All in pkg/ir.go or similar public file
- [ ] Internal validators in internal/

### compiler/ Module
- [ ] Exports: `Parser` interface
- [ ] Exports: `Compiler` interface
- [ ] Exports: `OptimizationPass` interface
- [ ] Exports: `CodeGenerator` interface
- [ ] Implementations in internal/
- [ ] All interfaces in pkg/compiler.go

### executors/ Module
- [ ] Exports: `ExecutorDriver` interface
- [ ] Exports: `SubmissionRequest`, `ExecutionStatus`
- [ ] Argo implementation in internal/argo/
- [ ] Airflow implementation in internal/airflow/
- [ ] Local implementation in internal/local/
- [ ] Registry pattern for driver discovery

### storage/ Module
- [ ] Exports: `Store` interface (Save/Get methods)
- [ ] Exports: `Query` interface (complex queries)
- [ ] Exports: `Transaction` interface
- [ ] PostgreSQL implementation in internal/postgres/
- [ ] Redis implementation in internal/redis/
- [ ] Independent (no dependencies on other modules)

### lineage/ Module
- [ ] Exports: `LineageEngine` interface
- [ ] Exports: `Graph` interface (Nodes, Edges, Queries)
- [ ] PostgreSQL backend in internal/postgres/
- [ ] In-memory graph in internal/graph/

### api/ Module
- [ ] Proto files in proto/src/flowforge/v1/
- [ ] Generated code in proto/gen/ (gitignored)
- [ ] gRPC services defined (PipelineService, ExecutionService, etc)
- [ ] REST routes via gRPC-Gateway
- [ ] Handlers in internal/handler/
- [ ] Services in internal/service/

### sdk/ Module
- [ ] Exports: `Pipeline` class
- [ ] Exports: `Task` class
- [ ] Exports: `Client` class (gRPC wrapper)
- [ ] CLI in flowforge/cli/
- [ ] Decorators in flowforge/decorators.py

### ui/ Module
- [ ] Exports: React components
- [ ] Services: REST client wrappers
- [ ] State management: Zustand/Redux
- [ ] Hooks: useExecution, useLineage, etc

---

## 🧪 Testing Structure

### Unit Tests
- [ ] ir/tests/unit/ (IR validation, builder)
- [ ] compiler/tests/unit/ (parser, optimizer, codegen)
- [ ] runtime/tests/unit/ (runner, executor)
- [ ] executors/tests/unit/ (driver implementations)
- [ ] storage/tests/unit/ (query builders)
- [ ] lineage/tests/unit/ (graph operations)
- [ ] api/tests/unit/ (handlers, middleware)
- [ ] sdk/tests/unit/ (pipeline, task, decorators)
- [ ] ui/tests/unit/ (components, hooks)
- [ ] Target: >80% coverage per module

### Integration Tests
- [ ] compiler/tests/integration/ (compiler → codegen)
- [ ] api/tests/integration/ (API → storage)
- [ ] sdk/tests/integration/ (SDK → API)
- [ ] ui/tests/integration/ (UI components with store)
- [ ] Target: >70% coverage

### E2E Tests
- [ ] tests/e2e/test_yaml_to_argo.py (YAML → Argo Workflow)
- [ ] tests/e2e/test_sdk_to_local.py (SDK → Local execution)
- [ ] tests/e2e/test_api_integration.py (Full API flow)
- [ ] Test fixtures in tests/fixtures/
- [ ] Sample pipelines in tests/fixtures/pipelines/
- [ ] Target: happy path + error cases

### Test Organization
- [ ] Each module: tests/{unit,integration,fixtures}
- [ ] E2E: tests/e2e/
- [ ] Fixtures co-located with tests
- [ ] conftest.py files for shared fixtures
- [ ] CI runs all tests on PR

---

## 📖 Documentation

### Root Level Documentation
- [ ] README.md (overview, quick start, commands)
- [ ] ARCHITECTURE.md (system design)
- [ ] MODULE_BOUNDARIES.md (ownership, interfaces)
- [ ] CONTRIBUTING.md (contribution guidelines)
- [ ] REPOSITORY_STRUCTURE.md (directory tree)

### Module Documentation
- [ ] Each module has README.md
- [ ] Each module README includes:
  - [ ] Purpose & responsibility
  - [ ] Key interfaces/exports
  - [ ] Dependencies (what does it depend on)
  - [ ] Usage examples
  - [ ] How to test locally

### Development Documentation
- [ ] docs/development/architecture-overview.md
- [ ] docs/development/module-responsibilities.md
- [ ] docs/development/adding-executor.md
- [ ] docs/development/testing-guide.md
- [ ] docs/development/contributing.md

### User Documentation
- [ ] docs/getting-started.md
- [ ] docs/user-guide/sdk-reference.md
- [ ] docs/user-guide/yaml-reference.md
- [ ] docs/user-guide/cli-reference.md

### API Documentation
- [ ] docs/api/grpc-api.md
- [ ] docs/api/rest-api.md (generated from proto)
- [ ] docs/api/proto-definitions.md

### Deployment Documentation
- [ ] docs/deployment/docker-setup.md
- [ ] docs/deployment/kubernetes-setup.md
- [ ] docs/deployment/terraform-deployment.md
- [ ] docs/deployment/helm-installation.md

---

## 🛠️ Build & Development Commands

### Makefile Tasks (Root Level)
- [ ] `make setup-dev` (setup environment)
- [ ] `make build` (build all modules)
- [ ] `make test` (all tests)
- [ ] `make test-unit` (unit tests only)
- [ ] `make test-integration` (integration tests)
- [ ] `make test-e2e` (E2E tests)
- [ ] `make lint` (lint all code)
- [ ] `make fmt` (format all code)
- [ ] `make lint-fix` (auto-fix issues)
- [ ] `make proto-gen` (generate protobuf)
- [ ] `make dev-up` (start services)
- [ ] `make dev-down` (stop services)
- [ ] `make build-images` (build Docker images)
- [ ] `make help` (show commands)

### Module Makefiles
- [ ] Each Go module has Makefile
- [ ] Each module supports: build, test, lint, fmt
- [ ] Python/Node modules use appropriate tools

### GitHub Actions CI/CD
- [ ] .github/workflows/ci-test.yml (run tests)
- [ ] .github/workflows/lint.yml (lint code)
- [ ] .github/workflows/build-images.yml (build on merge)
- [ ] .github/workflows/docs.yml (build docs)

### Docker & Compose
- [ ] docker-compose.yml (local dev: Postgres, Redis, etc)
- [ ] Dockerfile.api (API server image)
- [ ] Dockerfile.executor (executor image) [future]
- [ ] Dockerfile.ui (UI image) [future]

---

## 🔒 Code Quality Gates

### Before Merge
- [ ] All tests passing (unit, integration, E2E)
- [ ] Code coverage maintained/improved (>80% for new code)
- [ ] Linting passes (no warnings)
- [ ] Code formatted correctly
- [ ] No circular dependencies
- [ ] No internal/ imports across modules
- [ ] Documentation updated

### Reviewer Checklist
- [ ] Code follows module boundaries
- [ ] Public interfaces well-designed
- [ ] Tests adequate (unit + integration + E2E)
- [ ] Documentation clear
- [ ] Performance implications considered
- [ ] Security concerns addressed
- [ ] No code duplication
- [ ] Error handling appropriate

---

## 🚀 MVP Deliverables

### End of Week 8
- [ ] All 8 Go modules functional
- [ ] Python SDK usable (basic)
- [ ] React UI (basic dashboard)
- [ ] API server running (gRPC + REST)
- [ ] PostgreSQL schema migrated
- [ ] Argo driver working
- [ ] Local driver working
- [ ] Docker images built & pushed
- [ ] E2E tests passing
- [ ] Documentation complete
- [ ] Version: v0.1.0-alpha

---

## 🎯 Quality Metrics

| Metric | Target | Actual |
|--------|--------|--------|
| Go code coverage | >80% | __ |
| Python code coverage | >80% | __ |
| E2E tests | 100% pass | __ |
| Lint errors | 0 | __ |
| Circular dependencies | 0 | __ |
| Internal imports (cross-module) | 0 | __ |
| Documentation completeness | 100% | __ |
| Module tests pass locally | 100% | __ |
| CI/CD pipeline passing | 100% | __ |

---

## ✅ Deployment Readiness

- [ ] All modules build successfully
- [ ] All tests pass (CI/CD)
- [ ] Code scanned for vulnerabilities (SAST)
- [ ] Dependencies up-to-date
- [ ] Docker images scanned (Trivy)
- [ ] Terraform plan succeeds
- [ ] Helm charts validate
- [ ] Documentation complete & accurate
- [ ] Release notes written
- [ ] Changelog updated

---

## 📋 Launch Checklist

- [ ] GitHub repo public/internal
- [ ] Contributing guidelines reviewed by team
- [ ] License (Apache 2.0) in place
- [ ] Issue templates created
- [ ] PR template created
- [ ] Protected main branch
- [ ] Required status checks enabled
- [ ] Slack channel created
- [ ] Documentation site live
- [ ] First issue filed
- [ ] First PR reviewed
- [ ] Version tagged (v0.1.0-alpha)
- [ ] Announcement written

---

## 🐛 Known Issues / TBD

| Item | Status | Notes |
|------|--------|-------|
| Argo cluster setup | TBD | Local K8s (minikube/kind) |
| Airflow driver | Future | Phase 2 (after MVP) |
| Ray/Spark executors | Future | Phase 4 |
| Multi-tenancy | Future | Phase 3 |
| Lineage UI | Future | Phase 1 |
| Cost tracking | Future | Phase 1 |
| Self-healing | Future | Phase 3 |

---

## 📞 Sign-Off

- [ ] Architecture approved by Principal Architect
- [ ] All module owners assigned
- [ ] Team members allocated
- [ ] Development environment ready
- [ ] Timeline agreed upon (8 weeks for MVP)
- [ ] Budget approved

---

## 📝 Notes

**Completed By**: ________________________ **Date**: __________

**Verified By**: ________________________ **Date**: __________

**Approved By**: ________________________ **Date**: __________

