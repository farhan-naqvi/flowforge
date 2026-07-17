# FlowForge

**A local-first workbench that compiles one Python-authored workflow to
multiple orchestrators — and tells you exactly what each target preserves and
drops before you commit.**

Author a pipeline once as plain decorated Python. FlowForge parses it into a
canonical, versioned IR, validates the graph, compiles it to **Argo Workflows**
and **Apache Airflow**, reports the **capability/semantic-loss** for each target,
and runs it **locally for real** so you can iterate before touching a cluster.

```python
from flowforge import pipeline, task

@task(image="python:3.11")
def extract():
    return [{"id": i, "value": i * 10} for i in range(5)]

@task(retries=3, timeout="30s")
def transform(extract):                 # depends on `extract` by name
    return [{**r, "value": r["value"] * 2} for r in extract]

@task
def load(transform):
    print(f"loaded {len(transform)} rows")

etl = pipeline("simple-etl", tasks=[extract, transform, load], owner="data-team")
```

```
$ ff validate   examples/simple_etl.py     # graph + schema + injection-safe identifiers
$ ff capability examples/simple_etl.py      # what Argo / Airflow preserve vs drop
$ ff compile    examples/simple_etl.py argo # deterministic, safe Argo YAML
$ ff run        examples/simple_etl.py      # real local execution, real logs
$ ff serve      --workspace examples        # the workbench UI at http://127.0.0.1:8080
```

## Why FlowForge

Every orchestrator validates your pipeline only on its own server or cluster,
and no tool tells you *what you lose* moving a workflow from Airflow to Argo (or
vice-versa). FlowForge is the pre-deploy step: define once, validate in CI, and
get an honest per-target capability report — so you design and de-risk a
workflow before committing to any engine. It is **not** another orchestrator and
does not run your production workloads; mature schedulers already do that well.

See [docs/product/MARKET_AND_POSITIONING.md](docs/product/MARKET_AND_POSITIONING.md)
for the full positioning and competitive analysis.

## Feature status (honest)

| Capability | State |
|---|---|
| Author pipelines as decorated Python | **Working** |
| Canonical IR (`flowforge.io/v1`) + JSON | **Working** |
| Validation: DAG cycles/reachability, schema, injection-safe identifiers | **Working** |
| Per-target capability / semantic-loss report | **Working** |
| Compile to Argo Workflows (safe YAML) | **Working** |
| Compile to Apache Airflow (Airflow-3 style, safe emission) | **Working** |
| Real local execution + SQLite run/event store | **Working** |
| Workbench UI (JSON API + static UI, no build step) | **Working** |
| Go IR (`ir/`) as language-neutral schema-of-record | **Working** (builds + tested) |
| VS Code extension | **Experimental** — under `experimental/`, superseded by the CLI |
| Prefect/Dagster demo flows, FastAPI store | **Simulated/legacy** — under `experimental/` |
| Additional targets (Step Functions, Kestra…), scheduling, conditionals | **Planned** — see roadmap |
| Kubernetes/live submission, multi-tenancy, auth, sandboxing | **Not built** — explicit non-goals for the MVP |

Everything under [`experimental/`](experimental/README.md) is quarantined: not
built, tested, or supported.

## Install & run

Requires Python 3.11+ (tested on 3.11–3.13). Only runtime dependency is PyYAML.

```bash
pip install -e ".[dev]"      # or: make setup
ff --help
make demo                    # guided end-to-end walkthrough incl. a failure case
make test                    # 30 Python tests
make test-go                 # Go IR schema-of-record
```

## How it fits together

```
author (Python @task) ─▶ canonical IR (flowforge.io/v1) ─▶ validate
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        ▼                           ▼                           ▼
  capability report          compile: Argo YAML          run locally (real)
  (supported/lossy/            + Airflow DAG               → SQLite events
   unsupported per target)     (safe emission)             → UI / API
```

One implementation in [`flowforge/`](flowforge) backs every surface (CLI, API,
UI). Details in [docs/architecture/TARGET_ARCHITECTURE.md](docs/architecture/TARGET_ARCHITECTURE.md).

## Documentation

- [Repository audit](docs/audit/REPOSITORY_AUDIT.md) — verified starting state
- [Target architecture](docs/architecture/TARGET_ARCHITECTURE.md)
- [Decision log](docs/architecture/DECISION_LOG.md) — product & architecture ADRs
- [Market & positioning](docs/product/MARKET_AND_POSITIONING.md)
- [Threat model & security baseline](docs/security/THREAT_MODEL.md)
- [Demo script](docs/DEMO.md)
- [Roadmap](docs/ROADMAP.md) · [Known limitations](docs/KNOWN_LIMITATIONS.md)
- [Contributing](CONTRIBUTING.md)

## Security

FlowForge is a **local developer tool**. `ff run` executes your own pipeline
code in-process with **no sandboxing** — treat third-party pipeline files like
any script you would run. Compilation is injection-safe (validated identifiers +
`yaml.safe_dump`/`repr()` emission). Full model in
[docs/security/THREAT_MODEL.md](docs/security/THREAT_MODEL.md).

## License

[Apache License 2.0](LICENSE).
