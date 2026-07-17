# FlowForge Roadmap

Honest, evidence-based. Items are **Working** (shipped and tested today),
**Next**, or **Later**. Nothing here is marked done unless it runs in this repo.
See [KNOWN_LIMITATIONS.md](KNOWN_LIMITATIONS.md) for the boundary of "real."

## Working today

- Author pipelines as decorated Python (`@task` / `pipeline`).
- Canonical versioned IR (`flowforge.io/v1`), JSON in/out, Go schema-of-record (`ir/`).
- Validation: DAG cycles/reachability, schema, and injection-safe identifiers.
- Per-target capability / semantic-loss report (Argo, Airflow).
- Compile to Argo Workflows (safe YAML) and Apache Airflow (Airflow-3 style, safe emission).
- Real local execution with a SQLite run/event store.
- Workbench UI (JSON API + dependency-free static UI) and `ff` CLI.
- CI (Python matrix 3.11–3.13, Go IR, pip-audit/govulncheck advisory).

## Next (the next ~2 weeks)

- **Golden-file compile tests** checked into the repo (snapshot Argo YAML + Airflow
  DAG + capability report per example) so codegen changes are reviewable.
- **`ff diff`**: source-to-generated and target-to-target diff, surfacing exactly
  what changes between Argo and Airflow for one pipeline.
- **Schedule support** in the IR (`Schedule` task / cron tag) → Airflow `schedule=`
  and Argo `CronWorkflow`, with capability entries updated from `unsupported`.
- **VS Code extension as a thin client** of the CLI (delete the duplicate regex
  parser and duplicate codegen currently under `experimental/flowforge-vscode`).
- **DAG visualization** in the UI (render the IR graph; the audit noted a viable
  Cytoscape approach — vendored locally, no CDN).

## Later (the next ~6 weeks)

- **Conditional branching** in the IR → Argo `when:` and Airflow `BranchPythonOperator`,
  with capability diagnostics.
- **Additional compile targets**, opt-in and capability-gated: AWS Step Functions,
  Kestra. Each is a capability table entry + a compiler, not a rewrite.
- **`ff import`**: parse an existing Airflow DAG or Argo Workflow into the canonical
  IR (the "migration" use case, demoted from a product pillar to a feature).
- **Policy checks** on the IR (naming, resource ceilings, required owners) runnable
  in CI.
- **Headless/CI ergonomics**: machine-readable (`--json`) output for every command,
  non-zero exit codes wired for pipelines.

## Six-month architecture direction

- **Shared/versioned IR registry** and cross-environment **drift detection** —
  the natural first commercial layer on top of the OSS IR (see
  [MARKET_AND_POSITIONING.md](product/MARKET_AND_POSITIONING.md)).
- **OpenLineage emission** from generated pipelines so runtime lineage tools work
  with FlowForge-authored workflows.
- Represent **agent/LLM steps** in the IR *if and when* a durable standard settles —
  not as a marketing label, only if it solves a concrete orchestration problem
  better than existing tools.

## Explicit non-goals

Reimplementing a production scheduler; live Kubernetes/Argo submission or Airflow
scheduling from FlowForge; multi-tenancy; authentication/RBAC; sandboxed execution
of untrusted code. Mature engines already do the first; the rest are out of scope
for a local developer tool and would require the hardening in
[../docs/security/THREAT_MODEL.md](security/THREAT_MODEL.md).
