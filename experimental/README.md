# Experimental / quarantined code

Everything in this directory is **not part of the FlowForge product** and is
**not built, tested, or supported**. It was moved here during the repository
hardening (see [../docs/audit/REPOSITORY_AUDIT.md](../docs/audit/REPOSITORY_AUDIT.md)
and [ADR-003](../docs/architecture/DECISION_LOG.md)) because it does not
compile, is mock-only, or duplicates the canonical implementation.

The living product is the top-level [`flowforge/`](../flowforge) Python package.

| Path | State | Why it is here |
|---|---|---|
| `compiler/` | does not build | Go compiler: stub executors returning empty artifacts + a phantom IR (`task.Config.Image`) that doesn't match `ir/pkg`. Superseded by `flowforge/compilers`. |
| `runtime/` | mock only | In-memory container/registry mocks with hardcoded digests; no `go.mod`. Not on the MVP path. |
| `observability/` | mock only | In-memory metrics/logs/cost with a fabricated cost formula; superseded by `flowforge/store.py`. |
| `deployment/` | does not build | Terraform/Helm string generators with a receiver-type bug. Out of scope. |
| `executors/` | does not build | Argo/Airflow mock clients against a nonexistent IR. |
| `web/` | demo, fabricated data | Go demo server that returned hardcoded pipelines and fake execution metrics. Superseded by `flowforge/server.py`. |
| `demo/` | prints hardcoded JSON | Marketing demo. Superseded by `docs/DEMO.md`. |
| `ui/` | does not build | React fragments with no `package.json` calling endpoints that never existed. Superseded by `flowforge/webui/`. |
| `sdk/` | import broken | Python SDK: good authoring ideas but a broken decorator and a dependency on a nonexistent PyPI package. Ideas folded into `flowforge/`. |
| `flowforge-cli/` | works, superseded | The original working AST→codegen CLI. Its approach lives on (rebuilt, hardened) in `flowforge/`. |
| `flowforge-vscode/` | compiles, superseded | VS Code extension using regex parsing + duplicate codegen; should become a thin client of `flowforge/` (see roadmap). |
| `integrations/` | demo | FastAPI/SQLite store (real, minimal) + Prefect/Dagster demo flows with simulated workloads. Store concept folded into `flowforge/store.py`. |

Nothing here should be imported by the product or referenced by documentation as
a shipped capability.
