# FlowForge Repository Audit

Date: 2026-07-14 · Auditor: Lead Architect (multi-agent transformation) · Branch: `claude/flowforge-core-hardening`
Baseline commit: `70f5d8d` (main, clean tree at audit start)

Ground truth = code + reproduced execution. Documentation was treated as claims and verified.

## 1. Verification commands run (real results)

| Command | Result |
|---|---|
| `cd ir && go build ./...` | Core packages build; `tests/unit` FAILS (two package names `graph`+`ir` in one dir) |
| `cd ir && go test ./...` | FAIL — tests/unit setup failed; tests/integration build fails (`encoding/json` imported, unused). **Zero Go tests runnable** |
| `cd compiler && go build ./...` | FAIL — `pkg/interfaces.go`, `pkg/compiler.go`: `undefined: pkg` (imports `flowforge/ir/pkg` whose package name is `ir`, refers to it as `pkg`); executor subpackages reference IR fields that don't exist (`task.Config.Image`, string-typed `edge.From`) |
| `cd sdk && pytest tests` | FAIL at collection — `flowforge/decorators/common.py:10` `@task(image=...)` raises `TypeError`; also `pyproject.toml` depends on nonexistent PyPI package `flowforge-ir>=0.1.0` |
| `flowforge-cli`: `python -m flowforge.cli validate examples/simple_etl.py` | **WORKS** — validates |
| `flowforge-cli`: `generate argo examples/simple_etl.py` | **WORKS** — produces plausible Argo Workflow YAML (DAG deps, retries, activeDeadlineSeconds) |
| Go toolchain | Was not installed on this machine; no version pinning anywhere (go.work says 1.22; brew installed 1.26.5) |

## 2. Module inventory

| Path | Language | What it actually is | Builds? | Tests run? |
|---|---|---|---|---|
| `ir/` | Go | Clean IR structs, JSON Schema (`ir/spec.json`), DAG validator, builder, 4 JSON examples | yes (pkg) | **no** (broken test pkgs) |
| `compiler/` | Go | Pipeline orchestration skeleton + stub executors (`interfaces.go` TODOs return empty artifacts) + second set of "real" executors in `pkg/executors/` written against a **different, nonexistent IR** | **no** | no |
| `executors/argo`, `executors/airflow` | Go | Mock-client-based execution layer; imports `flowforge/ir` — **no go.mod anywhere in dir**; unbuildable | **no module** | no |
| `runtime/` | Go | Interface definitions + mock container runtime; no go.mod | **no module** | no |
| `observability/` | Go | In-memory metrics/logs/lineage/cost structs; no go.mod | **no module** | no |
| `deployment/` | Go | Terraform/Helm string generators + state mgmt; no go.mod | **no module** | no |
| `web/server.go` | Go | Single-file demo server: hardcoded pipeline; `/api/argo|airflow|terraform` return **hardcoded strings that ignore the pipeline**; `/api/execution` proxies localhost:8000 else **silently returns fabricated logs/metrics** ("45% CPU"); XSS via `innerHTML` of log text; CDN script | runs via `go run` (stdlib only) | n/a |
| `demo/main.go` | Go | Prints hardcoded pipeline JSON + capability marketing text | runs standalone | n/a |
| `sdk/` | Python | Real Pipeline/Task API, own IR mirror (`sdk/flowforge/ir/`), working-looking LocalExecutor (topo sort, in-process fn calls), CLI, visualizer | import broken | **no** (collection error) |
| `flowforge-cli/` | Python | **Best asset.** Real `ast`-based parser of `@pipeline`/`@task` Python → dataclass spec → validator → Argo YAML + Airflow DAG generators; click CLI; 2 examples with checked-in outputs | works | no tests exist |
| `flowforge-vscode/` | TypeScript | VS Code ext: **regex-based** Python parser (4+ RegExp), own DAG validator, own Argo+Airflow compilers (3rd+4th codegen implementations) | not verified | not verified |
| `ui/src/` | TypeScript | 4 React files (types, service, hook, editors). **No package.json** — cannot build. Calls `/api/compile`, `/api/validate`, `/api/estimate` — **no backend implements these** | **no** | no |
| `integrations/` | Python | FastAPI+SQLite observability API (real, minimal, no auth/limits); Prefect/Dagster demo flows (sleep-based fake workloads posting real events); `observability.db` **binary committed to git** | works | no |
| `docker-compose.yml` | — | Postgres/Redis/MinIO/Prometheus/Grafana/Loki/Jaeger with dev creds; mounts `./deployment/docker/*.yml` — **directory does not exist**; api/ui services commented out | partially broken | — |
| Root docs (25 files) | MD | Overlapping "COMPLETE"/"DELIVERY"/"SUMMARY" docs with unverifiable claims | — | — |

## 3. Critical findings

Severity: P0 blocks MVP · P1 major · P2 cleanup

| # | Sev | Finding | Evidence | Action |
|---|---|---|---|---|
| F1 | P0 | `compiler` Go module does not compile; shipped "compiler" is stubs returning empty artifacts | `compiler/pkg/interfaces.go:46` TODO; build errors above | Rebuild one canonical Go compiler path on `ir/pkg`, or demote Go compiler and canonicalize the working Python path |
| F2 | P0 | ≥6 incompatible pipeline models: `ir/pkg/spec.go`, phantom IR in `compiler/pkg/executors/*`, `sdk/flowforge/ir/spec.py`, `flowforge-cli` dataclasses, `ui/src/types/flowforge.ts`, vscode `pythonParser.ts`; plus hardcoded map in `web/server.go` | grep across files | One canonical versioned IR (JSON Schema in `ir/spec.json` is the best seed); all surfaces consume it |
| F3 | P0 | 4 independent Argo generators, 3 Airflow generators, none sharing code; the two hand-rolled Go YAML emitters drop env/resources and emit invalid `dependencies` as comma-string | `compiler/pkg/executors/argo/compiler.go:268-308`; `web/server.go:202` | Single codegen per target with golden tests |
| F4 | P0 | Fabricated execution presented as real: `/api/execution` fallback with fake metrics; no "simulated" labeling | `web/server.go:47-98,396-432` | Honest states: real events or clearly labeled sample mode |
| F5 | P0 | Build/test infrastructure is fiction: Makefile loops over nonexistent modules (`storage lineage api`), `tests/` and `docs/` dirs referenced but absent, no CI at all, no LICENSE despite Apache-2.0 claims | `Makefile:7`, `README.md:476` | Fix build path, add CI, add LICENSE |
| F6 | P0 | README claims "production-grade", "80%+ coverage", "30+ tests" — zero tests currently execute in Go, SDK tests can't collect | `README.md:11-14` | Truthful README |
| F7 | P1 | SDK depends on nonexistent `flowforge-ir` PyPI package; broken decorators module breaks all imports of `flowforge.decorators` | `sdk/pyproject.toml:27`, `sdk/flowforge/decorators/common.py:10` | Fix packaging + module |
| F8 | P1 | Security: no auth/rate/payload limits on observability API; XSS via log `innerHTML`; dev creds baked in compose (postgres flowforge/flowforge, minio minioadmin, grafana admin/admin); `latest` image tags; CDN script w/o SRI; committed SQLite db | `integrations/observability_api.py`, `web/server.go:583`, `docker-compose.yml:15,49,84` | Baseline hardening + THREAT_MODEL.md + .env.example |
| F9 | P1 | UI is two unrelated dead-ends: unbuildable React fragments calling phantom endpoints, and an embedded-HTML demo of hardcoded data | `ui/`, `web/server.go:434` | Pick one real UI for the MVP |
| F10 | P1 | VS Code extension parses Python with regexes and duplicates codegen — will diverge | `flowforge-vscode/src/parser/pythonParser.ts:30-92` | Delegate to CLI (single source of truth); mark experimental |
| F11 | P2 | 25 root markdown files with duplicated/conflicting status; placeholder org links (`your-org`); README self-contradicts (top: complete platform; bottom: "Implementation (in progress)") | root `*.md`; `README.md:465,496` | Consolidate to small doc set; archive the rest |
| F12 | P2 | Mock/in-memory Go subsystems (runtime, observability, deployment, executors) presented as platform components; interface design is decent but nothing is wired | dirs above | Quarantine as experimental design references or delete |

## 4. Hypothesis verdicts (from assignment §3)

All 14 preliminary hypotheses CONFIRMED except: "YAML or Python imports may use placeholder parsing" — **partially false**: `flowforge-cli` uses real `ast` parsing (vscode uses regex; there is no YAML authoring input anywhere despite README claiming it).

## 5. Assets worth keeping (verified working or sound)

1. `flowforge-cli` end-to-end path: ast parse → validate → Argo/Airflow codegen (works today).
2. `ir/spec.json` JSON Schema + `ir/pkg` Go structs/validator (clean, builds).
3. `integrations/observability_api.py` (small, real persistence; needs hardening).
4. SDK's `LocalExecutor` topo-sort in-process execution concept (needs repair).
5. `sdk` Pipeline/Task authoring API shape (decorator UX is good).
6. Cytoscape DAG visualization idea in web UI (needs honest data).

## 6. What FlowForge can honestly claim today

Author a pipeline as decorated Python → validate the DAG → generate Argo Workflows YAML and Airflow DAG code, via one working CLI. Everything else is design fiction, mocks, or demo fallback.
