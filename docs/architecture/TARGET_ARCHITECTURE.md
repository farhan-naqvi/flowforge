# FlowForge Target Architecture

Status: this describes the architecture we are building on `claude/flowforge-core-hardening`.
See [DECISION_LOG.md](DECISION_LOG.md) for the reasoning and [REPOSITORY_AUDIT.md](../audit/REPOSITORY_AUDIT.md) for the starting state.

## 1. One product, one path

```
             author (Python @pipeline/@task)
                        │
                        ▼
        ┌───────────────────────────────┐
        │  flowforge/ (canonical Python) │
        │                                │
        │  parse ──► IR (flowforge.io/v1)│   ◄── validated against ir/spec.json
        │            │                   │
        │            ├─ validate         │   graph (cycles, reachability, edges),
        │            │  (schema+safety+  │   identifiers (injection-safe),
        │            │   graph)          │   schema conformance
        │            │                   │
        │            ├─ capability check │   per target: supported / lossy / unsupported
        │            │                   │
        │            ├─ compile ─► Argo YAML   (safe emission)
        │            │           ─► Airflow DAG (safe emission)
        │            │                   │
        │            └─ run (local) ─► real in-process execution, topo order
        │                          └─► events ─► SQLite store
        └───────────────────────────────┘
              │            │            │
           CLI (ff)   HTTP+JSON API   static UI (honest states)
```

Every client surface (CLI, HTTP API, UI, and later VS Code/LSP) calls the **same** underlying functions in `flowforge/`. There is exactly one parser, one validator set, one capability engine, one compiler per target, one runner.

## 2. Canonical IR (`flowforge.io/v1`)

Contract: `ir/spec.json` (JSON Schema, draft-07). Language-neutral schema-of-record also implemented as Go structs in `ir/pkg` (builds + tested). Python is the working implementation.

Honestly-supported v1 fields and per-target mapping:

| IR field | Argo Workflows | Airflow (3.x-oriented) |
|---|---|---|
| task id, edges (task→task) | ✅ DAG `dependencies` | ✅ `>>` chaining |
| `handler.type` / `handler.source` | ✅ container `command` | ⚠️ operator choice is **lossy** (PythonOperator by default) |
| `retry.maxAttempts` | ✅ `retryStrategy.limit` | ✅ task `retries` |
| `retry.backoff/multiplier` | ✅ `retryStrategy.backoff` | ⚠️ **lossy** (Airflow retry_delay only) |
| `timeout` | ✅ per-template `activeDeadlineSeconds` | ✅ `execution_timeout` |
| `resources` (cpu/memory) | ✅ container `resources` | ⚠️ **lossy** unless KubernetesPodOperator |
| `env` | ✅ container `env` | ✅ operator `env`/`op_kwargs` |
| `schedule` (cron) | ⚠️ needs `CronWorkflow` — **unsupported in v1** | ✅ DAG `schedule` |
| `Conditional` branching | ❌ **unsupported v1** | ❌ **unsupported v1** |
| `costEstimate`, port JSON Schema | ℹ️ metadata only, not emitted | ℹ️ metadata only |

Anything marked ⚠️/❌ is surfaced by the **capability report** at compile time — never dropped silently.

## 3. Control plane vs execution plane

FlowForge does **not** reimplement a scheduler. It owns **authoring, validation, capability analysis, compilation, and a local dev runner**. Production execution belongs to Argo/Airflow (via generated artifacts) or, for local development, to the built-in in-process runner. This boundary is why the Go `runtime/`, `executors/`, and `deployment/` mock subsystems are quarantined — they attempted to reimplement execution and cluster provisioning that mature engines already provide.

## 4. Truthful product states

`source` (parsed) · `validated` · `invalid` (diagnostics) · `compiled` (artifact generated for a target) · `running` · `succeeded` · `failed` · `simulated` (explicitly labeled, never silent) · `stale` (source changed since last validate/compile/run).

## 5. Package layout (target)

```
flowforge/                 canonical Python package (the product)
  ir.py                    IR dataclasses + JSON (flowforge.io/v1)
  authoring.py             @pipeline / @task decorators
  parser.py                AST-based Python → IR
  validation.py            identifier safety + graph + schema conformance
  capability.py            per-target capability/loss report
  compilers/argo.py        IR → Argo Workflows YAML (safe emission)
  compilers/airflow.py     IR → Airflow DAG Python (safe emission)
  runner.py                real local in-process execution + events
  store.py                 SQLite run/task event store
  server.py                stdlib HTTP: JSON API + static UI
  webui/                   dependency-free honest UI (no build step)
  cli.py                   `ff` command
ir/                        Go schema-of-record (ir/spec.json + structs, tested)
experimental/              quarantined: Go compiler/runtime/deployment/observability/executors, React ui/
docs/                      consolidated documentation
tests/                     pytest: IR, parser, validation, golden compile, runner, API
```

## 6. Non-goals (this assignment)

Kubernetes/cluster provisioning · live Argo submission · live Airflow scheduling · multi-tenancy · auth/RBAC · sandboxed execution of untrusted code · Terraform/Helm/cost/lineage as product features (quarantined experimental) · additional compile targets beyond Argo + Airflow · agent/MCP workflows.
