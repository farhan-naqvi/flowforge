# FlowForge Architecture Decision Log

Format per entry: Decision · Context/Evidence · Alternatives rejected · Consequences/Risk.
Authoritative decisions by the Lead Architect after the Stage-1 audits and Stage-2 debate.

---

## ADR-001 — Product thesis: local-first portable workflow workbench

**Decision.** FlowForge becomes a **local-first developer workbench that compiles one Python-authored workflow to multiple orchestrators (Argo Workflows + Apache Airflow) and reports, per target, exactly what each preserves and drops.** The wedge is *pre-deploy validation + honest capability/semantic-loss diagnostics*, not lossless "run anywhere" portability.

**Context / evidence.**
- Only one path in the repo actually runs today: Python AST → validate → Argo/Airflow codegen (`flowforge-cli`). Audit §6.
- Market scan (July 2026): Prefect acquired Dagster 2026-07-13 — category consolidation sharpens buyer lock-in anxiety, but no incumbent owns *engine-agnostic pre-deploy validation*. Direction "workbench + compiler" scored highest (72) on the strategist's rubric.
- Novel, unclaimed sliver: a **capability report** telling you "Airflow preserves task retries; Argo maps timeout to `activeDeadlineSeconds`; neither v1 target supports conditional branching."

**Alternatives rejected.**
- *General-purpose orchestrator (Dir 1)* — suicidal competition (Airflow, Prefect/Dagster, Kestra, Temporal); no runtime in repo.
- *"LLVM/Terraform for workflows" framing* — technically unjustified: LLVM/Terraform own the execution substrate; a transpiler to someone else's runtime captures no lock-in. Prior art (Couler, CWL) shipped this exact thesis and stalled. We keep the compiler but **drop the analogy and the "run anywhere" promise**.
- *Agent/MCP orchestrator* — crowded (LangGraph/Temporal/Kestra), zero repo fit, undemonstrable here. Deferred: IR may represent agent steps later.
- *Multi-runtime platform (Dir 6)* — equals Kestra's/Prefect's stated claim; needs an engine we don't have.

**Consequences / risk.** Switching-pull is weak (a team on one orchestrator rarely needs a second). We mitigate by selling *design & de-risk before you commit* and by keeping adoption cost near zero (a CLI/LSP you add to CI). Strategist confidence 6/10 — success hinges on winning local DX first and never over-promising portability.

---

## ADR-002 — One canonical IR (`flowforge.io/v1`); Python is the single implementation

**Decision.** Adopt the existing `ir/spec.json` (JSON Schema, draft-07, `flowforge.io/v1`) as the **canonical, versioned IR contract**. Build **one** implementation of parse → IR → validate → capability-check → codegen, in **Python**, as the new top-level `flowforge/` package. The Go `ir/` module is kept as the **language-neutral schema-of-record** (it builds and its tests pass), not as a second implementation.

**Context / evidence.** Audit found ≥6 incompatible pipeline models (F2). The Go `compiler/` does not build (stubs + a phantom IR with `task.Config.Image`/string edges that don't exist). Repairing four Go executors + optimizer/validator buys nothing the working Python path lacks, and no Go toolchain was even installed. `ir/spec.json` is already dual-implemented (Go structs + Python dataclasses) and is the one artifact surfaces agree on in intent.

**Alternatives rejected.**
- *Repair the Go compiler, make Python a thin client* — large rewrite against a real IR, no runtime benefit; Go kept only as schema-of-record.
- *Keep the flat `flowforge-cli` model* — too poor for a credible portability story (no ports, no backoff policy, no per-target config); it silently drops information.

**Consequences / risk.** Risk of re-introducing the "spec richer than codegen" gap that made the Go path fiction. **Mitigation: IR v1 declares its honestly-supported subset, and the capability report emits `supported`/`lossy`/`unsupported` for everything a target can't faithfully map — richness never silently over-promises.**

---

## ADR-003 — Consolidate to one package; quarantine unbuildable/mocked code

**Decision.** Create canonical Python package **`flowforge/`** (authoring decorators + AST parser + IR + validators + capability engine + Argo/Airflow compilers + local runner + CLI + local server + static UI). Move non-viable code to **`experimental/`** with a README stating it does not build/run: Go `compiler/`, `runtime/`, `deployment/`, `observability/`, `executors/`, and the React `ui/` fragments. Keep `ir/` (Go schema-of-record). Fold the useful ideas from `flowforge-cli` and `sdk` into `flowforge/`, then archive those trees.

**Context / evidence.** Runtime audit: `runtime/`, `deployment/`, `observability/`, `executors/` are mock-only or don't compile (no `go.mod`; `HelmChartGenerator` receiver bug; phantom IR). UX audit: the 4 React files have no `package.json` and call endpoints that don't exist — unreachable code that invites "just fix the build."

**Alternatives rejected.** *Delete everything non-viable* — loses genuinely reusable design (executor codegen ideas, interface contracts). `experimental/` preserves them behind an honest boundary. *Leave them in place* — perpetuates the "looks like a platform" illusion the audit flagged.

**Consequences / risk.** Import paths and docs must stop referencing quarantined modules. Collaborator (Fable branch) may be editing some of these — see ADR-006.

---

## ADR-004 — Real local execution + honest run states; no fabricated data

**Decision.** The MVP executes user task functions **for real** (in-process, topological order) and persists real run/task events to SQLite. The UI/API distinguish `source → validated → invalid → compiled → running → succeeded → failed`, and any non-real data is explicitly labeled `SIMULATED`. Remove the silent fabricated-metrics fallback.

**Context / evidence.** Audit F4: `web/server.go` fabricates "45% CPU" logs with no label. Runtime audit: `sdk` `LocalExecutor` already does real topo-sort in-process execution (reusable once the import bug is fixed); `integrations/observability_api.py` schema is a fine event store.

**Consequences / risk.** **Executing user-authored Python in-process is inherent arbitrary code execution** — acceptable for a *local dev tool operating on your own code*, but the product must NEVER claim "secure execution of untrusted workflows" (see THREAT_MODEL). Trust boundary documented explicitly.

---

## ADR-005 — Security: identifier validation + escaped emission are in-scope for the MVP (not later)

**Decision.** Codegen must be **injection-safe now**. Enforce a DNS-1123-style identifier allowlist at IR-validation time for pipeline/task/owner names, and emit YAML via a real serializer and Python via safe rendering — never raw f-strings of user strings.

**Context / evidence.** Security audit reproduced **live RCE**: `@pipeline(name="etl'''; import os; os.system('id') #")` produced a valid malicious Airflow `dag_id=` line that Airflow's scheduler imports and executes. This is a shipping blocker for any "usable output" claim.

**Consequences / risk.** Some legal-in-Python identifiers become invalid pipeline names; acceptable and safer. Defense-in-depth (validate + escape) covers both the authored path and any future importer.

---

## ADR-006 — Collaboration: additive branch, quarantine over deletion

**Decision.** All work on `claude/flowforge-core-hardening`. Prefer **adding** the new `flowforge/` package and **moving** dead code to `experimental/` over rewriting files a concurrent Fable-5 collaborator might hold. Never force-push, reset, or merge the Fable branch. Keep commits small and atomic.

**Context / evidence.** Assignment §2. At audit start the tree was clean on `main` with no Fable commits visible; if Fable work appears it will most likely touch the same legacy Go/JS files, so we minimize edits there.

**Consequences / risk.** `git mv` of legacy trees can conflict if Fable edits them. We record an ownership map and, where a legacy file is only *referenced* (docs, Makefile), we rewrite the reference rather than the file.
