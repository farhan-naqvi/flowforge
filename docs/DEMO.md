# FlowForge Demo

A five-minute walkthrough of the real end-to-end path. Every command below runs
against code in this repository and produces real output — no fabricated data.

```bash
pip install -e ".[dev]"     # or: make setup
```

## One command

```bash
make demo          # runs scripts/demo.py: steps 1–8 below, incl. the failure case
```

## Or step by step

**1. Author** — a pipeline is plain decorated Python
([`examples/simple_etl.py`](../examples/simple_etl.py)): three `@task` functions
whose parameters name their upstreams.

**2. Canonical IR** — parsed statically, without executing your code:
```bash
ff ir examples/simple_etl.py
```
Emits `flowforge.io/v1` JSON: tasks (typed Source/Transform/Sink), edges, retry,
timeout, resources.

**3. Validate** — graph, schema, and injection-safe identifiers:
```bash
ff validate examples/simple_etl.py
# ✓ simple-etl: valid (3 tasks, 2 edges, 0 warnings)
```

**4. Capability report** — what each orchestrator preserves vs drops:
```bash
ff capability examples/simple_etl.py
```
For Airflow, `retry_backoff` shows `lossy` (only `retry_delay` is set); for both,
`conditional` shows `unsupported`. Nothing is dropped silently.

**5–6. Compile** — deterministic, injection-safe artifacts:
```bash
ff compile examples/simple_etl.py argo      # Argo Workflows YAML (via yaml.safe_dump)
ff compile examples/simple_etl.py airflow   # Airflow 3 DAG (repr-safe emission)
```
The Argo YAML parses with `yaml.safe_load`; the Airflow DAG passes `py_compile`.

**7. Run locally** — real in-process execution with real logs and durations,
persisted to a SQLite event store:
```bash
ff run examples/simple_etl.py
# ✓ extract    succeeded   … ms   extracted 5 rows
# ✓ transform  succeeded   … ms   transformed 5 rows
# ✓ load       succeeded   … ms   loaded 5 rows, total value 200
# succeeded  (stored in ~/.flowforge/runs.db)
```

**8. The failure case** — FlowForge rejects a broken pipeline instead of
silently "fixing" it ([`examples/broken_cycle.py`](../examples/broken_cycle.py),
a→b→a):
```bash
ff validate examples/broken_cycle.py
# error  [graph.cycle] pipeline contains a cycle; workflows must be acyclic
ff compile examples/broken_cycle.py argo
# ✗ refusing to compile broken-cycle: 1 validation error(s).
```

Try the same injection the pre-hardening code allowed — a pipeline named
`etl'''; import os; os.system('id') #` — and validation rejects it with
`name.unsafe` before any artifact is generated.

## The workbench UI

```bash
ff serve --workspace examples      # http://127.0.0.1:8080
```

Pick a pipeline to see its IR, diagnostics, the per-target capability matrix,
the generated artifacts (copy/download), and a **Run locally** button whose
results are labeled `LOCAL RUN` — never presented as remote/production execution.
