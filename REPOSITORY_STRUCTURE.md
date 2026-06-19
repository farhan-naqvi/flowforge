# FlowForge Repository Structure

## Project Tree

```
flowforge/
├── README.md
├── LICENSE (Apache 2.0)
├── CONTRIBUTING.md
├── ARCHITECTURE.md (root-level architecture overview)
├── Makefile (common tasks: build, test, lint, deploy)
│
├── api/                           # API Server (Go) & gRPC Definitions
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── cmd/
│   │   └── flowforge-server/
│   │       ├── main.go
│   │       └── config.go
│   ├── internal/
│   │   ├── server/
│   │   │   ├── api.go
│   │   │   ├── middleware.go
│   │   │   └── handlers/
│   │   │       ├── pipeline_handler.go
│   │   │       ├── execution_handler.go
│   │   │       ├── lineage_handler.go
│   │   │       └── cost_handler.go
│   │   ├── db/
│   │   │   ├── postgres.go
│   │   │   ├── migrations/
│   │   │   │   ├── 001_initial_schema.sql
│   │   │   │   └── 002_lineage_tables.sql
│   │   │   └── queries.go (SQL queries)
│   │   ├── auth/
│   │   │   ├── oidc.go
│   │   │   └── rbac.go
│   │   └── observability/
│   │       ├── metrics.go
│   │       ├── logging.go
│   │       └── tracing.go
│   ├── proto/                      # Protocol Buffers (gRPC definitions)
│   │   ├── flowforge/
│   │   │   ├── pipeline/
│   │   │   │   ├── v1/
│   │   │   │   │   ├── pipeline.proto
│   │   │   │   │   └── compilation.proto
│   │   │   ├── execution/
│   │   │   │   ├── v1/
│   │   │   │   │   ├── execution.proto
│   │   │   │   │   └── status.proto
│   │   │   ├── lineage/
│   │   │   │   └── v1/
│   │   │   │       └── lineage.proto
│   │   │   └── common/
│   │   │       ├── v1/
│   │   │       │   ├── ir.proto         # Core IR definitions
│   │   │       │   └── errors.proto
│   │   └── Makefile (protoc generation)
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── compiler_test.go
│   │   │   └── validator_test.go
│   │   ├── integration/
│   │   │   ├── api_test.go
│   │   │   └── fixtures/
│   │   └── testdata/
│   │       ├── sample_pipelines.yaml
│   │       └── fixtures.go
│   └── Dockerfile
│
├── compiler/                      # Compiler Core (Go)
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── ir/
│   │   │   ├── types.go             # IR data structures
│   │   │   ├── validator.go         # Validation rules
│   │   │   ├── builder.go           # IR construction
│   │   │   └── schema.go            # Schema contracts
│   │   ├── parser/
│   │   │   ├── parser.go            # Interface
│   │   │   ├── yaml/
│   │   │   │   └── parser.go
│   │   │   ├── sdk/
│   │   │   │   └── parser.go        # Parse SDK calls
│   │   │   └── ui/
│   │   │       └── parser.go        # Parse UI builder events
│   │   ├── compiler/
│   │   │   ├── compiler.go          # Main compiler logic
│   │   │   ├── optimizer/
│   │   │   │   ├── pass.go          # Optimization interface
│   │   │   │   ├── merge_tasks.go
│   │   │   │   └── parallelize.go
│   │   │   └── codegen/
│   │   │       ├── generator.go     # Codegen interface
│   │   │       ├── argo/
│   │   │       │   └── generator.go
│   │   │       ├── airflow/
│   │   │       │   └── generator.go
│   │   │       └── local/
│   │   │           └── generator.go
│   │   ├── lineage/
│   │   │   ├── engine.go
│   │   │   ├── graph.go
│   │   │   └── store.go             # PostgreSQL lineage store
│   │   └── cost/
│   │       ├── calculator.go
│   │       ├── estimator.go
│   │       └── models.go            # Cost models per executor
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── parser_test.go
│   │   │   ├── optimizer_test.go
│   │   │   ├── codegen_test.go
│   │   │   └── validator_test.go
│   │   ├── integration/
│   │   │   └── end_to_end_test.go
│   │   └── testdata/
│   │       ├── pipelines/
│   │       │   ├── simple.yaml
│   │       │   ├── complex_dag.yaml
│   │       │   └── with_contracts.yaml
│   │       └── expected_outputs/
│   │           ├── argo_workflow.yaml
│   │           └── airflow_dag.py
│   └── Makefile
│
├── executor/                      # Executor Drivers (Go)
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── driver/
│   │   │   ├── driver.go            # Driver interface
│   │   │   └── status.go
│   │   ├── argo/
│   │   │   ├── driver.go            # Argo implementation
│   │   │   ├── submitter.go
│   │   │   ├── monitor.go
│   │   │   └── error_handling.go
│   │   ├── local/
│   │   │   ├── driver.go            # Local implementation
│   │   │   ├── runner.go
│   │   │   ├── process.go
│   │   │   └── result_collector.go
│   │   └── airflow/
│   │       ├── driver.go            # Airflow implementation (future)
│   │       ├── dag_submitter.go
│   │       └── scheduler_monitor.go
│   ├── tests/
│   │   ├── unit/
│   │   │   └── driver_test.go
│   │   └── integration/
│   │       ├── argo_test.go
│   │       └── local_test.go
│   └── Makefile
│
├── sdk/                           # Python SDK
│   ├── setup.py
│   ├── pyproject.toml
│   ├── requirements.txt
│   ├── requirements-dev.txt
│   ├── flowforge/
│   │   ├── __init__.py
│   │   ├── client.py               # gRPC client
│   │   ├── pipeline.py             # Pipeline builder API
│   │   ├── task.py                 # Task definitions
│   │   ├── ir_builder.py           # IR construction helpers
│   │   ├── decorators.py           # @flowforge.task decorator
│   │   ├── compiler.py             # Local compiler wrapper
│   │   ├── cli/
│   │   │   ├── __init__.py
│   │   │   ├── main.py             # CLI entry point
│   │   │   ├── commands/
│   │   │   │   ├── init.py         # ff init
│   │   │   │   ├── local.py        # ff local
│   │   │   │   ├── submit.py       # ff submit
│   │   │   │   ├── status.py       # ff status
│   │   │   │   ├── logs.py         # ff logs
│   │   │   │   └── validate.py     # ff validate
│   │   │   └── config.py           # Config file parsing
│   │   └── validators/
│   │       ├── schema.py           # Schema validation
│   │       └── dag.py              # DAG cycle detection
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── test_pipeline.py
│   │   │   ├── test_task.py
│   │   │   └── test_ir_builder.py
│   │   ├── integration/
│   │   │   └── test_cli.py
│   │   └── fixtures/
│   │       └── conftest.py
│   └── Makefile
│
├── ui/                            # Frontend (React + TypeScript)
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── public/
│   ├── src/
│   │   ├── index.tsx
│   │   ├── App.tsx
│   │   ├── components/
│   │   │   ├── Builder/            # Pipeline builder
│   │   │   │   ├── Canvas.tsx
│   │   │   │   ├── TaskNode.tsx
│   │   │   │   ├── Toolbar.tsx
│   │   │   │   └── Properties.tsx
│   │   │   ├── Explorer/           # Lineage explorer
│   │   │   │   ├── LineageGraph.tsx
│   │   │   │   ├── TaskDetails.tsx
│   │   │   │   └── ReplayDialog.tsx
│   │   │   ├── Dashboard/          # Execution dashboard
│   │   │   │   ├── ExecutionList.tsx
│   │   │   │   ├── ExecutionDetail.tsx
│   │   │   │   ├── CostBreakdown.tsx
│   │   │   │   └── MetricsCharts.tsx
│   │   │   ├── Common/
│   │   │   │   ├── Header.tsx
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   └── Modals.tsx
│   │   │   └── Forms/
│   │   │       ├── PipelineForm.tsx
│   │   │       └── TaskForm.tsx
│   │   ├── services/
│   │   │   ├── api.ts              # REST client
│   │   │   ├── grpc-client.ts      # gRPC client
│   │   │   ├── pipeline.ts
│   │   │   ├── execution.ts
│   │   │   └── lineage.ts
│   │   ├── hooks/
│   │   │   ├── usePipeline.ts
│   │   │   ├── useExecution.ts
│   │   │   └── useLineage.ts
│   │   ├── store/                  # State management (Redux/Zustand)
│   │   │   ├── slices/
│   │   │   │   ├── pipelineSlice.ts
│   │   │   │   └── executionSlice.ts
│   │   │   └── index.ts
│   │   ├── pages/
│   │   │   ├── PipelineList.tsx
│   │   │   ├── PipelineEditor.tsx
│   │   │   ├── ExecutionDetail.tsx
│   │   │   └── NotFound.tsx
│   │   ├── types/
│   │   │   ├── api.ts              # API types (generated from proto)
│   │   │   └── domain.ts
│   │   └── styles/
│   │       └── index.css
│   ├── tests/
│   │   ├── unit/
│   │   │   └── components.test.tsx
│   │   └── e2e/
│   │       └── builder.e2e.ts
│   └── Dockerfile
│
├── infra/                         # Infrastructure as Code
│   ├── terraform/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── modules/
│   │   │   ├── kubernetes/
│   │   │   │   ├── main.tf
│   │   │   │   ├── variables.tf
│   │   │   │   └── outputs.tf
│   │   │   ├── database/
│   │   │   │   ├── main.tf
│   │   │   │   ├── variables.tf
│   │   │   │   └── outputs.tf
│   │   │   └── networking/
│   │   │       └── main.tf
│   │   └── environments/
│   │       ├── dev/
│   │       │   └── terraform.tfvars
│   │       ├── staging/
│   │       │   └── terraform.tfvars
│   │       └── prod/
│   │           └── terraform.tfvars
│   ├── helm/
│   │   ├── flowforge/
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   ├── values-dev.yaml
│   │   │   ├── values-prod.yaml
│   │   │   ├── templates/
│   │   │   │   ├── api-server-deployment.yaml
│   │   │   │   ├── api-server-service.yaml
│   │   │   │   ├── api-server-hpa.yaml
│   │   │   │   ├── postgres-statefulset.yaml
│   │   │   │   ├── redis-deployment.yaml
│   │   │   │   ├── prometheus-config.yaml
│   │   │   │   ├── grafana-deployment.yaml
│   │   │   │   └── rbac.yaml
│   │   │   └── dependencies/
│   │   │       ├── argo-workflows/
│   │   │       └── kube-prometheus/
│   │   └── scripts/
│   │       ├── install.sh
│   │       └── upgrade.sh
│   └── docker/
│       ├── Dockerfile.api
│       ├── Dockerfile.executor
│       └── docker-compose.yml      # Local dev environment
│
├── examples/                      # Example Pipelines
│   ├── README.md
│   ├── basic/
│   │   ├── extract_transform_load.yaml
│   │   └── extract_transform_load.py
│   ├── data_quality/
│   │   ├── pipeline.yaml
│   │   └── transformations/
│   │       └── validate_schema.py
│   ├── ml_workflow/
│   │   ├── pipeline.yaml
│   │   └── transformations/
│   │       ├── preprocess.py
│   │       ├── train.py
│   │       └── evaluate.py
│   └── multi_executor/
│       ├── argo_pipeline.yaml
│       └── airflow_pipeline.yaml
│
├── tests/                         # Cross-module e2e tests
│   ├── e2e/
│   │   ├── conftest.py
│   │   ├── test_yaml_to_argo.py    # YAML → Argo workflow
│   │   ├── test_sdk_to_local.py    # SDK → Local execution
│   │   ├── test_lineage_tracking.py
│   │   └── test_cost_estimation.py
│   └── fixtures/
│       ├── pipelines/
│       └── expected_outputs/
│
├── docs/                          # Documentation
│   ├── README.md
│   ├── getting-started.md
│   ├── user-guide/
│   │   ├── pipeline-basics.md
│   │   ├── sdk-reference.md
│   │   ├── yaml-reference.md
│   │   ├── ui-builder.md
│   │   ├── execution.md
│   │   ├── replay.md
│   │   ├── lineage.md
│   │   └── cost-tracking.md
│   ├── development/
│   │   ├── architecture.md           # Links to root ARCHITECTURE.md
│   │   ├── contributing.md
│   │   ├── module-guide/
│   │   │   ├── ir.md
│   │   │   ├── compiler.md
│   │   │   ├── executor-drivers.md
│   │   │   ├── api-server.md
│   │   │   └── ui.md
│   │   ├── adding-executor.md       # How to add new executor
│   │   └── testing.md
│   ├── deployment/
│   │   ├── docker-compose-setup.md
│   │   ├── kubernetes-setup.md
│   │   ├── terraform-deployment.md
│   │   ├── helm-installation.md
│   │   └── security.md
│   ├── api/
│   │   ├── grpc-api.md              # gRPC service definitions
│   │   └── rest-api.md              # REST endpoint docs
│   └── troubleshooting.md
│
├── scripts/                       # Development scripts
│   ├── setup-dev.sh                # Local dev environment setup
│   ├── run-tests.sh                # Test runner
│   ├── build-all.sh                # Build all components
│   ├── proto-gen.sh                # Generate protobuf code
│   └── docker-build.sh             # Build Docker images
│
├── .github/
│   ├── workflows/
│   │   ├── test.yml                # Run tests on PR
│   │   ├── lint.yml                # Lint on PR
│   │   ├── build.yml               # Build images on merge
│   │   └── docs.yml                # Build docs site
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug.md
│   │   └── feature.md
│   └── PULL_REQUEST_TEMPLATE.md
│
├── .gitignore
├── .dockerignore
├── Makefile                       # Root-level tasks
├── docker-compose.yml             # Local dev environment
├── go.mod (root workspace)
├── go.sum (root workspace)
├── requirements-all.txt           # Python dependencies (all)
└── VERSION                        # Version tag

```

## Directory Organization Principles

1. **Top-level by concerns**: `api/`, `compiler/`, `executor/`, `sdk/`, `ui/`, `infra/`
2. **Go modules**: Each service is a Go module (`go.mod`)
3. **Internal packages**: `internal/` for private packages (Go convention)
4. **Proto files**: Centralized in `api/proto/` for single source of truth
5. **Tests co-located**: Tests next to source code (Go: `_test.go`, Python: `test_*.py`)
6. **Cross-module tests**: `tests/e2e/` for integration tests
7. **Examples**: Real-world pipeline examples in `examples/`
8. **Infrastructure**: Terraform + Helm templates in `infra/`
9. **Documentation**: Comprehensive guide for users & developers in `docs/`

