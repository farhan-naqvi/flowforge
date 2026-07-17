# Covenant

**Contract-first, verify-before-run data pipelines.** Define ODCS source + target
contracts, describe a pipeline as structured intent, let the planner compose it
from the `covenant-transforms` primitive library, and Covenant checks it against
the target contract — **statically** (schema inference, before any data moves)
and **dynamically** (a local DuckDB run on synthetic + edge-case data) — before
compiling it to an Argo/Spark run.

> The moat is **trust** (verify-before-run), not the code generation. See
> [../docs/covenant/VISION_AND_ARCHITECTURE.md](../docs/covenant/VISION_AND_ARCHITECTURE.md).

## Quick start

```bash
pip install -e covenant-transforms
pip install -e covenant                       # pulls Ibis + DuckDB
covenant demo                                 # end-to-end, incl. a failure case
covenant serve                                # workbench UI at http://127.0.0.1:8090
```

## The loop

```
ODCS contracts ──▶ intent ──▶ planner ──▶ static schema-inference conformance
                                              │  (rejects a plan that can't
                                              │   produce the target schema)
                                              ▼
                        synthetic test data ─▶ local DuckDB verify ─▶ verdict
                                              │  (schema + quality rules:
                                              │   PK unique, not-null, non-empty)
                                              ▼
                         green? ──▶ compile Argo/Spark artifact ──▶ (one-click run)
```

## CLI

```bash
covenant plan    <intent.yaml>            # compose a plan; show static conformance
covenant verify  <intent.yaml>            # static + local dynamic verification
covenant compile <intent.yaml>            # Argo/Spark artifact (refuses non-conformant plans)
covenant demo                             # the sales/daily_orders example, end to end
covenant serve                            # workbench UI + JSON API (loopback only)
```

## What's built vs deferred

Built and tested: the primitive library (Ibis), the trust engine (Slices 0–1),
the workbench UI + GitOps plan write (Slice 2), and Argo/Spark compile (Slice 3).
Deferred with clean interfaces (not faked): an LLM-backed planner (the
deterministic planner ships today; an agent would propose steps validated by the
same verifier) and a live Spark/Argo cluster run (the artifact is generated;
running it needs a cluster). Details in the vision doc §10.

## No-code authoring (in the workbench)

`covenant serve` is a full authoring UI, not a read-only demo. You can:

- **Create a data product** (`domain/data_product`) — writes skeleton ODCS
  contracts + an empty intent into `covenant-contracts/`.
- **Edit the source and target contracts** as structured fields: add, remove,
  rename, and reorder fields; change data types; toggle nullable/required; set or
  clear primary keys.
- **Edit the pipeline intent**: add/remove/reorder transform steps and their
  parameters.
- **Save** — atomic writes into the local GitOps tree (review and commit via a
  pull request; the tool never pushes or deploys).
- **Re-validate and verify immediately** after saving — the same planner/verify
  pipeline the CLI uses. Invalid contracts cannot be saved without a visible
  explanation.

## Security

Local single-user tool. `covenant serve` binds `127.0.0.1`, rejects non-loopback
Host headers, caps payloads, and runs single-threaded (the embedded DuckDB used
by verify is not safe under concurrent request threads). It reads and writes only
inside the local `covenant-contracts` tree; it performs no Git push, remote
access, or deployment.
