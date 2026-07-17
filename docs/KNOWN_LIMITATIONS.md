# Known Limitations

FlowForge is an early (alpha) local developer tool. This page is the honest
boundary of what it does and does not do, so you can trust the parts that work.

## Feature reality

| Area | Status | Notes |
|---|---|---|
| Python authoring, IR, validation, capability report, Argo + Airflow codegen, local run, UI/CLI | **Working** | Covered by 30 tests + Go IR tests in CI |
| Compile targets | **Argo + Airflow only** | Step Functions, Kestra, etc. are on the roadmap, not built |
| Scheduling (cron) & conditional branching | **Not emitted in IR v1** | Reported as `unsupported` in the capability report rather than dropped silently |
| VS Code extension | **Experimental** | Under `experimental/`; uses regex parsing and duplicates codegen; not the product path |
| Prefect/Dagster demos, FastAPI store, docker-compose stack | **Legacy / simulated** | Under `experimental/`; demo workloads are simulated (sleep-based), not real pipelines |
| Go `compiler/`, `runtime/`, `deployment/`, `observability/`, `executors/`, React `ui/` | **Quarantined, do not build** | Moved to `experimental/`; superseded by the `flowforge/` package. See [experimental/README.md](../experimental/README.md) |

## Execution

- **`ff run` executes your pipeline code in-process with no sandboxing.** It is a
  local dev tool. Do not run untrusted/third-party pipeline files without reading
  them first. See [security/THREAT_MODEL.md](security/THREAT_MODEL.md).
- The local runner executes tasks **in topological order, single-process**. There
  is no parallelism, no retry execution (retries are compiled into the target
  artifacts, not applied by the local runner), and no distributed execution.
- FlowForge does **not** submit to Kubernetes/Argo or an Airflow scheduler. It
  generates artifacts you deploy yourself.

## Compilation fidelity

- Generated **Airflow** DAGs use `PythonOperator` stubs that raise
  `NotImplementedError` — they are a faithful DAG skeleton (ids, deps, retries,
  timeouts, schedule) you wire to your task implementations, not a drop-in runnable
  module. This is intentional and honest: FlowForge cannot know your deployment's
  imports.
- Generated **Argo** workflows reference a generic `flowforge_task` entrypoint via
  container `args`; you supply the image/entrypoint for your environment.
- Fields the target cannot represent faithfully are reported as `lossy` or
  `unsupported` (e.g. Airflow retry backoff is `lossy`; conditional branching is
  `unsupported` on both). Nothing is dropped without a diagnostic.

## Server / storage

- The workbench server is for **local single-user** use. It binds `127.0.0.1` by
  default, has **no authentication**, and its `/api/run` endpoint executes code.
- The SQLite store is guarded by a process-wide lock — fine for local use, not for
  concurrent multi-writer deployments. Log fields are capped (64 KB) to bound growth.

## Platform / versions

- Python 3.11–3.13 (tested). Only runtime dependency: PyYAML.
- Go 1.22+ for the `ir/` schema-of-record (the product does not require Go at runtime).

## Claims we deliberately do NOT make

"Production-grade", "enterprise-ready", "secure execution of untrusted workflows",
"multi-tenant", any measured-coverage percentage beyond what CI reports, or "runs
your pipelines" (it compiles and locally runs them; it does not operate your
production orchestrator).
