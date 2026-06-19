# FlowForge Monorepo Structure

## Directory Tree with Ownership & Interfaces

```
flowforge/
├── .github/                          # GitHub Actions, issue templates
│   ├── workflows/
│   │   ├── ci-test.yml              # Run all tests
│   │   ├── lint.yml                 # Lint all modules
│   │   ├── build-images.yml         # Build Docker images on merge
│   │   └── docs.yml                 # Generate docs site
│   ├── PULL_REQUEST_TEMPLATE.md
│   └── ISSUE_TEMPLATE/
│       ├── bug.md
│       ├── feature.md
│       └── architecture.md
│
├── ir/                              # Intermediate Representation (Core)
│   ├── README.md
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── spec/
│   │   │   ├── pipeline.go          # PipelineSpec definition
│   │   │   ├── task.go              # TaskSpec definition
│   │   │   ├── edge.go              # DAG edge definition
│   │   │   ├── policy.go            # ExecutionPolicy definition
│   │   │   ├── schema.go            # SchemaContract definition
│   │   │   └── types.go             # Common types (ResourceRequirement, etc)
│   │   ├── validator/
│   │   │   ├── validator.go         # Validator interface & orchestrator
│   │   │   ├── schema_validator.go  # Validate spec format
│   │   │   ├── dag_validator.go     # Detect cycles, validate DAG structure
│   │   │   ├── resource_validator.go # Validate resource constraints
│   │   │   └── contract_validator.go # Validate schema contracts
│   │   ├── builder/
│   │   │   ├── builder.go           # IRBuilder for programmatic construction
│   │   │   └── helpers.go           # Helper methods
│   │   └── serialize/
│   │       ├── json.go              # IR ↔ JSON serialization
│   │       ├── protobuf.go          # IR ↔ Protobuf serialization
│   │       └── yaml.go              # IR ↔ YAML serialization
│   ├── pkg/                         # Public API
│   │   ├── ir.go                    # Main interface definitions
│   │   └── errors.go                # Error types
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── validator_test.go
│   │   │   ├── builder_test.go
│   │   │   └── serialize_test.go
│   │   ├── fixtures/
│   │   │   ├── valid_pipelines.go
│   │   │   └── invalid_pipelines.go
│   │   └── testdata/
│   │       ├── simple.ir.json
│   │       ├── complex_dag.ir.json
│   │       └── with_contracts.ir.json
│   └── Makefile
│
├── sdk/                             # Python SDK + CLI
│   ├── README.md
│   ├── Makefile
│   ├── setup.py
│   ├── pyproject.toml
│   ├── requirements.txt
│   ├── requirements-dev.txt
│   ├── flowforge/
│   │   ├── __init__.py
│   │   ├── client.py                # gRPC/REST client (wraps API)
│   │   ├── pipeline.py              # Pipeline builder (fluent API)
│   │   ├── task.py                  # Task definition
│   │   ├── decorators.py            # @flowforge.task decorator
│   │   ├── ir_builder.py            # IR construction helpers
│   │   ├── validators.py            # Local validation (wraps ir/validator)
│   │   ├── cli/
│   │   │   ├── __init__.py
│   │   │   ├── main.py              # CLI entry point (ff command)
│   │   │   ├── commands/
│   │   │   │   ├── init.py          # ff init (create project)
│   │   │   │   ├── local.py         # ff local pipeline.yaml
│   │   │   │   ├── submit.py        # ff submit --executor argo
│   │   │   │   ├── status.py        # ff status <execution-id>
│   │   │   │   ├── logs.py          # ff logs <execution-id>
│   │   │   │   └── validate.py      # ff validate pipeline.yaml
│   │   │   └── config.py            # CLI config file parsing
│   │   └── utils/
│   │       ├── validation.py        # Validation helpers
│   │       └── logger.py            # Structured logging
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── test_pipeline.py
│   │   │   ├── test_task.py
│   │   │   ├── test_decorators.py
│   │   │   └── test_validators.py
│   │   ├── integration/
│   │   │   ├── test_cli.py
│   │   │   └── test_client.py
│   │   └── fixtures/
│   │       └── conftest.py
│   └── dist/                        # Built wheels/distributions
│
├── compiler/                        # Compiler (IR → Executor Config)
│   ├── README.md
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── compiler/
│   │   │   ├── compiler.go          # Main compiler orchestrator
│   │   │   ├── parser.go            # Parser interface
│   │   │   ├── optimizer.go         # Optimizer interface & chain
│   │   │   └── codegen.go           # CodeGenerator interface & registry
│   │   ├── parser/
│   │   │   ├── yaml/
│   │   │   │   ├── parser.go        # YAML parser impl
│   │   │   │   └── decoder.go       # YAML decoding helpers
│   │   │   ├── sdk/
│   │   │   │   ├── parser.go        # SDK event parser (from Python)
│   │   │   │   └── event_handler.go
│   │   │   └── builder/
│   │   │       ├── parser.go        # IR builder event parser
│   │   │       └── event_handler.go
│   │   ├── optimizer/
│   │   │   ├── pass.go              # OptimizationPass interface
│   │   │   ├── registry.go          # Pass registry
│   │   │   ├── merge_tasks.go       # Sequential task merge
│   │   │   ├── parallelize.go       # Parallelism detection
│   │   │   └── resource_pool.go     # Shared resource pools
│   │   ├── codegen/
│   │   │   ├── registry.go          # CodeGenerator registry
│   │   │   ├── argo/
│   │   │   │   ├── generator.go     # Argo YAML generation
│   │   │   │   ├── workflow.go      # Argo Workflow spec builder
│   │   │   │   └── template.go      # Argo template helpers
│   │   │   ├── airflow/
│   │   │   │   ├── generator.go     # Airflow DAG generation
│   │   │   │   ├── dag_builder.go   # Airflow DAG spec builder
│   │   │   │   └── operator.go      # Airflow operator mappings
│   │   │   └── local/
│   │   │       ├── generator.go     # Local DAG generation
│   │   │       ├── dag.go           # Go DAG struct definition
│   │   │       └── executor_config.go # Local executor config
│   │   └── validation/
│   │       ├── compiler_validator.go # Compiler-specific validation
│   │       └── rules.go             # Custom validation rules
│   ├── pkg/
│   │   ├── compiler.go              # Public API
│   │   └── errors.go
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── parser_test.go
│   │   │   ├── optimizer_test.go
│   │   │   └── codegen_test.go
│   │   ├── integration/
│   │   │   ├── end_to_end_test.go
│   │   │   └── fixtures.go
│   │   └── testdata/
│   │       ├── pipelines/
│   │       │   ├── simple.yaml
│   │       │   ├── complex_dag.yaml
│   │       │   └── with_contracts.yaml
│   │       └── expected_outputs/
│   │           ├── argo/
│   │           │   └── simple.yaml
│   │           ├── airflow/
│   │           │   └── simple.py
│   │           └── local/
│   │               └── simple.json
│   └── Makefile
│
├── runtime/                         # Runtime Execution (Local dev mode)
│   ├── README.md
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── runner/
│   │   │   ├── runner.go            # Local runtime orchestrator
│   │   │   ├── executor.go          # Task executor interface
│   │   │   ├── process_executor.go  # Subprocess task executor
│   │   │   └── dag_runner.go        # DAG dependency resolver & executor
│   │   ├── process/
│   │   │   ├── subprocess.go        # Subprocess management
│   │   │   ├── logger.go            # Stdout/stderr capture
│   │   │   └── signal.go            # Signal handling (SIGTERM, etc)
│   │   ├── storage/
│   │   │   ├── cache.go             # Local execution cache
│   │   │   └── result_store.go      # Store execution results locally
│   │   └── observability/
│   │       ├── metrics.go           # Local runtime metrics
│   │       └── logger.go            # Local execution logging
│   ├── pkg/
│   │   ├── runtime.go               # Public API
│   │   └── errors.go
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── runner_test.go
│   │   │   ├── executor_test.go
│   │   │   └── dag_runner_test.go
│   │   └── integration/
│   │       └── e2e_test.go
│   └── Makefile
│
├── executors/                       # Executor Drivers (Argo, Airflow, etc)
│   ├── README.md
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── driver/
│   │   │   ├── driver.go            # ExecutorDriver interface
│   │   │   ├── registry.go          # Driver registry
│   │   │   ├── status.go            # Status types (pending, running, etc)
│   │   │   └── errors.go            # Executor-specific errors
│   │   ├── argo/
│   │   │   ├── driver.go            # Argo driver implementation
│   │   │   ├── submitter.go         # Submit workflow to Argo
│   │   │   ├── monitor.go           # Monitor workflow status
│   │   │   ├── logger.go            # Fetch workflow logs
│   │   │   ├── error_handler.go     # Argo-specific error handling
│   │   │   └── config.go            # Argo configuration
│   │   ├── airflow/
│   │   │   ├── driver.go            # Airflow driver implementation
│   │   │   ├── submitter.go         # Submit DAG to Airflow
│   │   │   ├── monitor.go           # Monitor DAG run status
│   │   │   ├── logger.go            # Fetch task logs from Airflow
│   │   │   ├── error_handler.go     # Airflow-specific error handling
│   │   │   └── config.go            # Airflow configuration
│   │   └── local/
│   │       ├── driver.go            # Local driver (wraps runtime)
│   │       └── config.go
│   ├── pkg/
│   │   ├── executor.go              # Public API
│   │   └── errors.go
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── argo_test.go
│   │   │   └── airflow_test.go
│   │   └── integration/
│   │       ├── argo_integration_test.go
│   │       └── airflow_integration_test.go
│   └── Makefile
│
├── lineage/                         # Lineage Tracking & Provenance
│   ├── README.md
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── engine/
│   │   │   ├── engine.go            # Lineage engine orchestrator
│   │   │   ├── graph.go             # In-memory lineage graph
│   │   │   ├── builder.go           # Build lineage from execution
│   │   │   └── query.go             # Query lineage graph
│   │   ├── store/
│   │   │   ├── store.go             # LineageStore interface
│   │   │   ├── postgres.go          # PostgreSQL implementation
│   │   │   ├── schema.go            # Database schema
│   │   │   └── migrations.go        # Database migrations
│   │   ├── models/
│   │   │   ├── node.go              # Lineage node (task, dataset)
│   │   │   ├── edge.go              # Lineage edge (data flow)
│   │   │   └── metadata.go          # Provenance metadata
│   │   └── observability/
│   │       ├── metrics.go           # Lineage metrics
│   │       └── logger.go            # Lineage logging
│   ├── pkg/
│   │   ├── lineage.go               # Public API
│   │   └── errors.go
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── engine_test.go
│   │   │   ├── graph_test.go
│   │   │   └── query_test.go
│   │   ├── integration/
│   │   │   └── store_test.go
│   │   └── testdata/
│   │       └── sample_lineage.json
│   └── Makefile
│
├── storage/                         # Data Storage Layer (PostgreSQL, Redis)
│   ├── README.md
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── postgres/
│   │   │   ├── client.go            # PostgreSQL client
│   │   │   ├── connection.go        # Connection pooling
│   │   │   ├── migrations/
│   │   │   │   ├── 001_initial_schema.sql
│   │   │   │   ├── 002_executions_table.sql
│   │   │   │   ├── 003_lineage_tables.sql
│   │   │   │   └── migrations.go
│   │   │   ├── queries/
│   │   │   │   ├── pipeline_queries.go
│   │   │   │   ├── execution_queries.go
│   │   │   │   ├── lineage_queries.go
│   │   │   │   └── cost_queries.go
│   │   │   └── transaction.go       # Transaction management
│   │   ├── redis/
│   │   │   ├── client.go            # Redis client
│   │   │   ├── cache.go             # Caching layer
│   │   │   └── queue.go             # Job queue
│   │   ├── models/
│   │   │   ├── pipeline.go          # Pipeline model
│   │   │   ├── execution.go         # Execution model
│   │   │   ├── task_run.go          # Task run model
│   │   │   └── cost.go              # Cost model
│   │   └── health/
│   │       └── checker.go           # Health check queries
│   ├── pkg/
│   │   ├── store.go                 # Public API
│   │   └── errors.go
│   ├── tests/
│   │   ├── unit/
│   │   │   └── queries_test.go
│   │   └── integration/
│   │       └── postgres_test.go
│   └── Makefile
│
├── api/                             # API Server (gRPC + REST)
│   ├── README.md
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── flowforge-api/
│   │       └── main.go              # API server entry point
│   ├── internal/
│   │   ├── server/
│   │   │   ├── server.go            # gRPC server setup
│   │   │   ├── grpc_server.go       # gRPC service implementations
│   │   │   └── http_server.go       # REST middleware (gRPC-Gateway)
│   │   ├── handler/
│   │   │   ├── pipeline_handler.go  # Pipeline CRUD handlers
│   │   │   ├── execution_handler.go # Execution trigger/status handlers
│   │   │   ├── lineage_handler.go   # Lineage query handlers
│   │   │   └── cost_handler.go      # Cost tracking handlers
│   │   ├── service/
│   │   │   ├── pipeline_service.go  # Pipeline business logic
│   │   │   ├── execution_service.go # Execution orchestration
│   │   │   ├── lineage_service.go   # Lineage business logic
│   │   │   └── cost_service.go      # Cost calculation
│   │   ├── middleware/
│   │   │   ├── auth.go              # Authentication middleware
│   │   │   ├── rbac.go              # RBAC authorization
│   │   │   ├── logging.go           # Request/response logging
│   │   │   ├── metrics.go           # Prometheus metrics
│   │   │   └── errors.go            # Error handling
│   │   ├── config/
│   │   │   ├── config.go            # Config parsing
│   │   │   └── validation.go        # Config validation
│   │   └── observability/
│   │       ├── metrics.go           # API metrics
│   │       ├── logging.go           # Structured logging
│   │       └── tracing.go           # Distributed tracing
│   ├── proto/                       # gRPC Protocol Buffers
│   │   ├── gen/                     # Generated Go code (gitignore)
│   │   └── src/
│   │       ├── flowforge/
│   │       │   ├── v1/
│   │       │   │   ├── pipeline/
│   │       │   │   │   ├── pipeline.proto
│   │       │   │   │   └── service.proto
│   │       │   │   ├── execution/
│   │       │   │   │   ├── execution.proto
│   │       │   │   │   └── service.proto
│   │       │   │   ├── lineage/
│   │       │   │   │   ├── lineage.proto
│   │       │   │   │   └── service.proto
│   │       │   │   ├── common/
│   │       │   │   │   ├── ir.proto
│   │       │   │   │   └── errors.proto
│   │       │   │   └── cost/
│   │       │   │       ├── cost.proto
│   │       │   │       └── service.proto
│   │       │   └── buf.yaml          # Buf package management
│   ├── pkg/
│   │   ├── api.go                   # Public API
│   │   └── errors.go
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── handler_test.go
│   │   │   ├── middleware_test.go
│   │   │   └── service_test.go
│   │   ├── integration/
│   │   │   └── api_test.go
│   │   └── fixtures/
│   │       └── sample_requests.go
│   ├── Dockerfile
│   └── Makefile
│
├── observability/                   # Metrics, Logging, Tracing
│   ├── README.md
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── metrics/
│   │   │   ├── registry.go          # Prometheus metric registry
│   │   │   ├── collector.go         # Metric collector interface
│   │   │   ├── api_metrics.go       # API-specific metrics
│   │   │   ├── compiler_metrics.go  # Compiler metrics
│   │   │   ├── execution_metrics.go # Execution metrics
│   │   │   └── storage_metrics.go   # Storage metrics
│   │   ├── logging/
│   │   │   ├── logger.go            # Structured JSON logger
│   │   │   ├── fields.go            # Log field helpers
│   │   │   └── middleware.go        # HTTP logging middleware
│   │   ├── tracing/
│   │   │   ├── tracer.go            # OTEL tracer initialization
│   │   │   ├── span_processor.go    # Span processor config
│   │   │   └── instrumentation.go   # OTEL instrumentation
│   │   └── health/
│   │       ├── checker.go           # Health check orchestrator
│   │       └── probes.go            # Liveness/readiness probes
│   ├── pkg/
│   │   ├── observability.go         # Public API
│   │   └── errors.go
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── metrics_test.go
│   │   │   └── logger_test.go
│   │   └── integration/
│   │       └── observability_test.go
│   └── Makefile
│
├── deployment/                      # Infrastructure & Deployment
│   ├── README.md
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
│   │   │   ├── networking/
│   │   │   │   ├── main.tf
│   │   │   │   └── variables.tf
│   │   │   └── observability/
│   │   │       ├── main.tf
│   │   │       ├── prometheus.tf
│   │   │       └── grafana.tf
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
│   │   │   │   ├── api-server/
│   │   │   │   │   ├── deployment.yaml
│   │   │   │   │   ├── service.yaml
│   │   │   │   │   ├── hpa.yaml
│   │   │   │   │   ├── configmap.yaml
│   │   │   │   │   └── secret.yaml
│   │   │   │   ├── postgres/
│   │   │   │   │   ├── statefulset.yaml
│   │   │   │   │   ├── service.yaml
│   │   │   │   │   └── pvc.yaml
│   │   │   │   ├── redis/
│   │   │   │   │   ├── deployment.yaml
│   │   │   │   │   └── service.yaml
│   │   │   │   ├── observability/
│   │   │   │   │   ├── prometheus.yaml
│   │   │   │   │   ├── grafana.yaml
│   │   │   │   │   ├── loki.yaml
│   │   │   │   │   └── jaeger.yaml
│   │   │   │   ├── rbac.yaml
│   │   │   │   ├── namespace.yaml
│   │   │   │   └── ingress.yaml
│   │   │   └── dependencies/
│   │   │       ├── argo-workflows/
│   │   │       └── kube-prometheus/
│   │   └── scripts/
│   │       ├── install.sh
│   │       ├── upgrade.sh
│   │       └── uninstall.sh
│   ├── docker/
│   │   ├── Dockerfile.api
│   │   ├── Dockerfile.executor
│   │   └── docker-compose.yml       # Local dev environment
│   └── scripts/
│       ├── build-images.sh
│       ├── push-images.sh
│       └── setup-cluster.sh
│
├── ui/                              # Frontend (React + TypeScript)
│   ├── README.md
│   ├── package.json
│   ├── package-lock.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── vitest.config.ts
│   ├── public/
│   │   ├── index.html
│   │   └── favicon.ico
│   ├── src/
│   │   ├── index.tsx
│   │   ├── App.tsx
│   │   ├── main.css
│   │   ├── components/
│   │   │   ├── Builder/
│   │   │   │   ├── Canvas.tsx       # DAG canvas
│   │   │   │   ├── TaskNode.tsx     # Task node component
│   │   │   │   ├── Toolbar.tsx      # Builder toolbar
│   │   │   │   ├── Properties.tsx   # Task properties panel
│   │   │   │   └── Builder.tsx      # Builder orchestrator
│   │   │   ├── Explorer/
│   │   │   │   ├── LineageGraph.tsx # Lineage visualization
│   │   │   │   ├── TaskDetails.tsx  # Task detail panel
│   │   │   │   ├── ReplayDialog.tsx # Replay controls
│   │   │   │   └── Explorer.tsx     # Explorer orchestrator
│   │   │   ├── Dashboard/
│   │   │   │   ├── ExecutionList.tsx # Execution list
│   │   │   │   ├── ExecutionDetail.tsx # Execution detail
│   │   │   │   ├── CostBreakdown.tsx  # Cost visualization
│   │   │   │   ├── MetricsCharts.tsx  # Metrics/performance charts
│   │   │   │   └── Dashboard.tsx    # Dashboard orchestrator
│   │   │   ├── Common/
│   │   │   │   ├── Header.tsx
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   ├── Modal.tsx
│   │   │   │   └── ErrorBoundary.tsx
│   │   │   └── Forms/
│   │   │       ├── PipelineForm.tsx
│   │   │       ├── TaskForm.tsx
│   │   │       └── SchemaForm.tsx
│   │   ├── services/
│   │   │   ├── api.ts               # REST client
│   │   │   ├── grpc-client.ts       # gRPC-web client
│   │   │   ├── pipeline.ts          # Pipeline API client
│   │   │   ├── execution.ts         # Execution API client
│   │   │   ├── lineage.ts           # Lineage API client
│   │   │   └── cost.ts              # Cost API client
│   │   ├── hooks/
│   │   │   ├── usePipeline.ts
│   │   │   ├── useExecution.ts
│   │   │   ├── useLineage.ts
│   │   │   └── useAuth.ts
│   │   ├── store/
│   │   │   ├── index.ts             # Store setup (Zustand/Redux)
│   │   │   ├── slices/
│   │   │   │   ├── pipelineSlice.ts
│   │   │   │   ├── executionSlice.ts
│   │   │   │   └── uiSlice.ts
│   │   │   └── middleware.ts        # Store middleware
│   │   ├── pages/
│   │   │   ├── PipelineList.tsx
│   │   │   ├── PipelineEditor.tsx
│   │   │   ├── ExecutionDetail.tsx
│   │   │   ├── Dashboard.tsx
│   │   │   └── NotFound.tsx
│   │   ├── types/
│   │   │   ├── api.ts               # API types (generated from proto)
│   │   │   ├── domain.ts            # Domain types
│   │   │   └── ui.ts                # UI state types
│   │   ├── utils/
│   │   │   ├── formatting.ts
│   │   │   ├── validation.ts
│   │   │   └── auth.ts
│   │   └── styles/
│   │       ├── variables.css
│   │       ├── base.css
│   │       └── components.css
│   ├── tests/
│   │   ├── unit/
│   │   │   ├── components.test.tsx
│   │   │   ├── hooks.test.tsx
│   │   │   └── store.test.ts
│   │   ├── integration/
│   │   │   └── workflows.test.tsx
│   │   ├── e2e/
│   │   │   ├── builder.e2e.ts
│   │   │   └── dashboard.e2e.ts
│   │   └── fixtures/
│   │       └── factories.ts
│   ├── Dockerfile
│   └── Makefile
│
├── examples/                        # Example Pipelines
│   ├── README.md
│   ├── basic/
│   │   ├── README.md
│   │   ├── etl_pipeline.yaml
│   │   ├── etl_pipeline.py
│   │   └── data/
│   │       └── sample_input.csv
│   ├── data_quality/
│   │   ├── README.md
│   │   ├── pipeline.yaml
│   │   ├── pipeline.py
│   │   └── transformations/
│   │       ├── validate_schema.py
│   │       ├── detect_anomalies.py
│   │       └── generate_report.py
│   ├── ml_workflow/
│   │   ├── README.md
│   │   ├── pipeline.yaml
│   │   ├── pipeline.py
│   │   └── transformations/
│   │       ├── preprocess.py
│   │       ├── train.py
│   │       ├── evaluate.py
│   │       └── requirements.txt
│   └── multi_executor/
│       ├── README.md
│       ├── argo_pipeline.yaml
│       ├── airflow_pipeline.yaml
│       └── local_pipeline.yaml
│
├── tests/                           # Cross-module e2e Tests
│   ├── README.md
│   ├── e2e/
│   │   ├── conftest.py              # Pytest fixtures
│   │   ├── test_yaml_to_argo.py     # YAML → Argo Workflow
│   │   ├── test_yaml_to_airflow.py  # YAML → Airflow DAG
│   │   ├── test_sdk_to_local.py     # SDK → Local execution
│   │   ├── test_sdk_to_argo.py      # SDK → Argo submission
│   │   ├── test_lineage_tracking.py # Execution → Lineage
│   │   ├── test_cost_tracking.py    # Cost estimation/tracking
│   │   ├── test_api_integration.py  # Full API flow
│   │   └── test_ui_integration.py   # UI builder flow
│   └── fixtures/
│       ├── pipelines/
│       │   ├── simple.yaml
│       │   ├── complex_dag.yaml
│       │   └── with_contracts.yaml
│       ├── expected_outputs/
│       │   ├── argo/
│       │   ├── airflow/
│       │   └── local/
│       └── docker-compose.test.yml  # Test environment
│
├── docs/                            # Documentation
│   ├── README.md                    # Docs index
│   ├── getting-started.md
│   ├── user-guide/
│   │   ├── pipeline-basics.md
│   │   ├── sdk-reference.md
│   │   ├── yaml-reference.md
│   │   ├── ui-builder.md
│   │   ├── execution-modes.md
│   │   ├── replay-diff.md
│   │   ├── lineage-tracking.md
│   │   └── cost-tracking.md
│   ├── development/
│   │   ├── architecture-overview.md
│   │   ├── module-responsibilities.md
│   │   ├── contributing.md
│   │   ├── testing-guide.md
│   │   └── adding-executor.md
│   ├── deployment/
│   │   ├── docker-setup.md
│   │   ├── kubernetes-setup.md
│   │   ├── terraform-deployment.md
│   │   ├── helm-installation.md
│   │   └── security-hardening.md
│   ├── api/
│   │   ├── grpc-api.md
│   │   ├── rest-api.md
│   │   └── proto-definitions.md
│   └── troubleshooting.md
│
├── scripts/                         # Development Scripts
│   ├── setup-dev.sh                 # Local dev environment
│   ├── run-tests.sh                 # Run all tests
│   ├── run-tests-unit.sh            # Run unit tests only
│   ├── run-tests-integration.sh     # Run integration tests
│   ├── run-tests-e2e.sh             # Run e2e tests
│   ├── build-all.sh                 # Build all modules
│   ├── build-images.sh              # Build Docker images
│   ├── proto-gen.sh                 # Generate protobuf code
│   ├── coverage-report.sh           # Generate coverage reports
│   ├── lint-all.sh                  # Lint all code
│   ├── fmt-all.sh                   # Format all code
│   └── local-cluster-up.sh          # Start local K8s + Argo
│
├── .github/                         # GitHub Actions & Templates
│   ├── workflows/
│   │   ├── ci-test.yml
│   │   ├── lint.yml
│   │   ├── build-images.yml
│   │   ├── docs.yml
│   │   └── security-scan.yml
│   ├── PULL_REQUEST_TEMPLATE.md
│   └── ISSUE_TEMPLATE/
│       ├── bug.md
│       ├── feature.md
│       └── architecture.md
│
├── .gitignore
├── .dockerignore
├── .editorconfig
├── LICENSE (Apache 2.0)
├── README.md                        # Root README
├── CONTRIBUTING.md
├── ARCHITECTURE.md                  # (Already created)
├── REPOSITORY_STRUCTURE.md          # (Already created)
├── DEPENDENCY_DIAGRAM.md            # (Already created)
├── MVP_AND_ROADMAP.md               # (Already created)
├── DESIGN_SUMMARY.md                # (Already created)
├── MODULE_BOUNDARIES.md             # (New - created below)
├── Makefile                         # Root-level tasks
├── docker-compose.yml               # Local dev environment
├── go.work                          # Go workspace (monorepo)
├── go.work.sum
├── requirements-all.txt             # All Python dependencies
├── pyproject.toml                   # Python workspace config
└── VERSION                          # Version tag (e.g., 0.1.0)
```

---

## Module Ownership & Responsibilities

| Module | Owner | Responsibility | Dependencies | Key Interfaces |
|--------|-------|-----------------|--------------|-----------------|
| **ir/** | Core Team | IR definition, validation, serialization | None (foundational) | `PipelineSpec`, `Validator`, `IRBuilder` |
| **compiler/** | Compiler Team | Parse, validate, optimize, codegen | `ir/` | `Parser`, `OptimizationPass`, `CodeGenerator` |
| **runtime/** | Execution Team | Local task execution, DAG runner | `ir/`, `compiler/` | `Runner`, `Executor`, `TaskExecutor` |
| **executors/** | Execution Team | Argo, Airflow, Local drivers | `ir/`, `compiler/` | `ExecutorDriver`, `SubmissionRequest`, `ExecutionStatus` |
| **storage/** | Data Team | PostgreSQL, Redis, queries, migrations | None (standalone) | `Store`, `Query`, `Transaction` |
| **lineage/** | Data Team | Lineage tracking, provenance, queries | `ir/`, `storage/` | `LineageEngine`, `LineageStore`, `Graph` |
| **api/** | Platform Team | gRPC + REST server, handlers, services | All core modules | gRPC services, REST routes |
| **observability/** | Platform Team | Metrics, logging, tracing | None (cross-cutting) | `MetricsCollector`, `Logger`, `Tracer` |
| **deployment/** | DevOps Team | Terraform, Helm, Docker, CI/CD | None (infrastructure) | Terraform modules, Helm values |
| **sdk/** | SDK Team | Python SDK, CLI, user-facing API | `ir/`, `compiler/`, `api/` | `Pipeline`, `Task`, `Client` |
| **ui/** | UI Team | React dashboard, builder, explorer | `api/` (REST client) | React components, state management |
| **tests/** | QA Team | E2E tests, fixtures, test orchestration | All modules | Test utilities, fixtures |

---

## Module Interfaces & Contracts

### Core Interfaces (Minimal)

#### IR Module - `ir/pkg/ir.go`
```go
// PipelineSpec defines a pipeline
type PipelineSpec interface {
    ID() string
    Name() string
    Version() string
    Tasks() []TaskSpec
    Edges() []Edge
    GetTask(id string) (TaskSpec, error)
}

// Validator validates IR
type Validator interface {
    Validate(ctx context.Context, spec PipelineSpec) []ValidationError
}

// IRBuilder constructs IR programmatically
type IRBuilder interface {
    AddTask(spec TaskSpec) IRBuilder
    Connect(from, to string, output, input string) IRBuilder
    Build() (PipelineSpec, error)
}
```

#### Compiler Module - `compiler/pkg/compiler.go`
```go
// Parser converts input → IR
type Parser interface {
    Parse(ctx context.Context, input interface{}) (ir.PipelineSpec, error)
    Supports(format string) bool
}

// OptimizationPass transforms IR
type OptimizationPass interface {
    Optimize(ctx context.Context, pipeline ir.PipelineSpec) (ir.PipelineSpec, error)
    Name() string
    AppliesTo(executor string) bool
}

// CodeGenerator produces executor config
type CodeGenerator interface {
    Generate(ctx context.Context, pipeline ir.PipelineSpec) (interface{}, error)
    ExecutorType() string
}

// Compiler orchestrates compilation
type Compiler interface {
    Compile(ctx context.Context, spec ir.PipelineSpec, executor string) (interface{}, error)
}
```

#### Executors Module - `executors/pkg/executor.go`
```go
// ExecutorDriver submits & monitors execution
type ExecutorDriver interface {
    Submit(ctx context.Context, config *SubmissionRequest) (*Submission, error)
    Status(ctx context.Context, submissionID string) (*ExecutionStatus, error)
    Logs(ctx context.Context, submissionID, taskID string) (io.Reader, error)
    Cancel(ctx context.Context, submissionID string) error
}

// ExecutionStatus tracks task progress
type ExecutionStatus struct {
    SubmissionID string
    State        ExecutionState // Pending, Running, Success, Failed
    Tasks        map[string]TaskStatus
    StartTime    time.Time
    EndTime      time.Time
}
```

#### Storage Module - `storage/pkg/store.go`
```go
// Store persists pipelines, executions, lineage
type Store interface {
    SavePipeline(ctx context.Context, spec ir.PipelineSpec) error
    GetPipeline(ctx context.Context, id string) (ir.PipelineSpec, error)
    SaveExecution(ctx context.Context, exec *Execution) error
    GetExecution(ctx context.Context, id string) (*Execution, error)
}

// Query interface for complex queries
type Query interface {
    GetExecutionHistory(ctx context.Context, pipelineID string, limit int) ([]*Execution, error)
    GetTaskMetrics(ctx context.Context, taskID string) (*TaskMetrics, error)
}
```

#### Lineage Module - `lineage/pkg/lineage.go`
```go
// LineageEngine tracks data provenance
type LineageEngine interface {
    RecordExecution(ctx context.Context, execution *Execution) error
    QueryLineage(ctx context.Context, dataID string) (*Graph, error)
}

// Graph represents lineage (tasks, data, edges)
type Graph interface {
    Nodes() []Node
    Edges() []Edge
    Upstream(nodeID string) []Node
    Downstream(nodeID string) []Node
}
```

#### API Module - `api/proto/src/flowforge/v1/*.proto`
```protobuf
service PipelineService {
    rpc CreatePipeline(CreatePipelineRequest) returns (CreatePipelineResponse);
    rpc GetPipeline(GetPipelineRequest) returns (GetPipelineResponse);
    rpc ListPipelines(ListPipelinesRequest) returns (ListPipelinesResponse);
    rpc UpdatePipeline(UpdatePipelineRequest) returns (UpdatePipelineResponse);
    rpc DeletePipeline(DeletePipelineRequest) returns (DeletePipelineResponse);
}

service ExecutionService {
    rpc SubmitExecution(SubmitExecutionRequest) returns (SubmitExecutionResponse);
    rpc GetExecutionStatus(GetExecutionStatusRequest) returns (GetExecutionStatusResponse);
    rpc GetExecutionLogs(GetExecutionLogsRequest) returns (stream ExecutionLog);
    rpc CancelExecution(CancelExecutionRequest) returns (CancelExecutionResponse);
}
```

---

## Dependency Graph

```
                    SDK (Python)
                         │ gRPC
                    API Server
                         │
        ┌────────────────┼────────────────┐
        │                │                │
    Compiler        Storage (DB)     Observability
        │                │                │
    ┌───┴────┐           │                │
    │         │           │                │
 Parser   Executors   Lineage Engine    (Cross-cutting)
    │      (Argo,
    │      Airflow,
    │      Local)
    │
    IR
(Foundation)

Module-level dependencies:
- sdk/ → api/ (gRPC client)
- api/ → compiler/, storage/, lineage/, observability/
- compiler/ → ir/
- executors/ → ir/, compiler/
- lineage/ → ir/, storage/
- runtime/ → ir/, compiler/
- ui/ → api/ (REST client)
- tests/ → all modules

Forbidden dependencies (acyclic):
- ir/ must not depend on compiler/, executors/, api/
- compiler/ must not depend on api/, storage/
- executors/ must not depend on api/
- storage/ only depends on itself (standalone)
```

---

## Module Boundaries & Contracts

### Boundary 1: IR ↔ Compiler
**Contract**: Compiler reads IR (immutable), produces executor config  
**Data Flow**: `ir.PipelineSpec` → `Compiler.Compile()` → executor-specific config (YAML, DAG, etc)  
**Boundary Enforcement**:
- IR defines immutable spec
- Compiler must not modify IR
- Compiler produces new data structures (not modify IR)

### Boundary 2: Compiler ↔ Executors
**Contract**: Executors receive compiled config, execute, return status  
**Data Flow**: `ExecutorConfig` → `ExecutorDriver.Submit()` → `Execution`  
**Boundary Enforcement**:
- Executors implement `ExecutorDriver` interface
- No direct coupling to specific executors (registry pattern)
- Status types standardized across executors

### Boundary 3: Executors ↔ Runtime (Local only)
**Contract**: Local executor uses runtime for in-process execution  
**Data Flow**: `ExecutorDriver` → `Runner.Execute()` → results  
**Boundary Enforcement**:
- Runtime only used by local executor
- Other executors (Argo, Airflow) use external systems

### Boundary 4: Execution ↔ Lineage
**Contract**: Execution publishes events, Lineage consumes  
**Data Flow**: `Execution` event → `LineageEngine.RecordExecution()` → stored graph  
**Boundary Enforcement**:
- Lineage engine subscribes to execution events
- Lineage does not drive execution (observer pattern)

### Boundary 5: All Modules ↔ Storage
**Contract**: All modules use Storage for persistence  
**Data Flow**: All → `Store.Save*()` / `Store.Get*()`  
**Boundary Enforcement**:
- Storage is independent (no dependencies on other modules)
- Single source of truth for all persistent state
- Migrations managed centrally

### Boundary 6: All Modules ↔ Observability
**Contract**: All modules emit metrics, logs, traces  
**Data Flow**: Any module → `MetricsCollector.Record()`, `Logger.Info()`, etc  
**Boundary Enforcement**:
- Observability is cross-cutting (injected via middleware)
- No module dependencies on observability
- Observability can be disabled (no-op implementations)

### Boundary 7: SDK ↔ API
**Contract**: SDK sends gRPC requests to API  
**Data Flow**: `Pipeline` (SDK) → gRPC → API → handlers  
**Boundary Enforcement**:
- SDK only knows API contract (gRPC)
- SDK unaware of backend implementation
- API versioned (v1, v2, etc)

### Boundary 8: UI ↔ API
**Contract**: UI sends REST requests to API  
**Data Flow**: UI component → REST → API handler  
**Boundary Enforcement**:
- UI only knows API contract (REST)
- UI generated from API (OpenAPI)
- API versioned in URL path

---

## No Cross-Module Implementation Dependencies

**Principle**: Modules depend only on interfaces, not implementations.

Example: Compiler doesn't import `executors/internal/argo/driver.go`
```go
// ✓ Correct: depend on interface
import "flowforge/executors/pkg"
driver := registry.GetDriver("argo")

// ✗ Wrong: direct dependency on implementation
import "flowforge/executors/internal/argo"
driver := &argo.Driver{}
```

---

## Communication Patterns

### Synchronous
- API ↔ Client (gRPC/REST)
- Compiler ↔ CodeGenerator (function calls)
- Parser ↔ IR (function calls)

### Asynchronous
- Execution → Lineage (events)
- Execution → Observability (metrics publish)
- SDK → API (potentially batched)

### Event-Driven
- Execution lifecycle: submitted → running → completed
- Lineage updates on task completion
- Metrics emitted on all state transitions

