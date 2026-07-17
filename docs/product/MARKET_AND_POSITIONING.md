# FlowForge — Market & Positioning

Date: July 2026. Sources are primary (vendor docs, release notes, announcements)
and listed at the end. This informs [ADR-001](../architecture/DECISION_LOG.md).

## Positioning statement

> **FlowForge is the local-first workbench where you define a pipeline once in
> Python, validate it against a canonical workflow IR in CI, and compile it to
> Airflow or Argo — with an explicit report of what each target preserves and
> drops — so teams design and de-risk workflows before committing to any
> orchestrator.**

It is **not** another orchestrator and does not run production workloads.

## Market context

The biggest recent event is **Prefect's acquisition of Dagster (announced
2026-07-13)** — the largest consolidation in the "modern orchestrator" tier.
Both open-source projects continue, but the category is concentrating and buyers
are more sensitive than ever to engine lock-in. Meanwhile Airflow 3 is the
enterprise default, Argo owns Kubernetes-native DAGs, and Kestra/Flyte/Temporal
each hold distinct niches. No incumbent owns **engine-agnostic, pre-deploy
validation**, and the only prior attempts at portable workflow compilation
(Couler, CWL) stalled.

## Competitor matrix

Overlap/gap are relative to FlowForge's honest asset: Python → canonical IR →
DAG validation → Argo YAML + Airflow DAG codegen + capability diagnostics + local run.

| Competitor | Positioning | Primary user | Differentiator | Overlap w/ FF | Gap FF can exploit |
|---|---|---|---|---|---|
| Airflow 3 | Enterprise standard scheduler | Platform/data teams | Assets, DAG versioning, hiring pool | FF *emits* Airflow DAGs | Weak local dev; single engine |
| Argo Workflows | K8s-native container DAGs | Platform/infra | Runs on Kubernetes | FF *emits* Argo YAML | Verbose YAML authoring/validation |
| Prefect (+Dagster) | Durable execution + declarative assets | Data/ML/AI | Now owns two models + MCP | Author→run | Consolidation raises lock-in fear |
| Kestra 1.0 | Declarative YAML control plane | Enterprise ops/data | NL→YAML, event-driven | Declarative authoring | Own engine only; not portable |
| Flyte 2.0 / Union | K8s-native ML + agentic | ML/AI eng | Massive fan-out | Little (runtime) | K8s barrier; not authoring-portable |
| Temporal | Durable execution, code-as-workflow | Backend/platform | Crash-proof long-running code | Almost none | Overkill for batch; no static DAG gate |
| Hamilton (DagWorks) | Lightweight Python dataflow | Data/ML eng | Runs anywhere Python runs | Closest philosophy | In-process only; no engine codegen/validation |
| Mage | Block-based pipelines | Small teams | Visual editor | Authoring UX | Not portable; SMB scope |
| Managed cloud (MWAA, Step Functions, Composer) | Hosted orchestration | Cloud-committed | Zero-ops | FF could *target* these | Per-cloud lock-in; migration pain |
| Couler / CWL (prior art) | Multi-engine compiler / portable standard | ML / bioinformatics | The exact "compile to many engines" idea | Same thesis | Both stalled — leaves a diagnostic gap unfilled |

## Market gaps

1. **Pre-deploy, local validation is universally weak** — every engine validates
   only on its own server/cluster; there is no engine-agnostic local + CI gate.
2. **Multi-engine shops maintain drifting duplicate definitions** — teams running
   both Argo and Airflow hand-maintain two specs. (This repository itself
   exhibited that drift — a demoable proof of the pain.)
3. **Portability/anti-lock-in has no credible tool, and no one reports the loss** —
   nobody tells you *what you give up* compiling to Argo vs Airflow.

## Target personas

- **Platform/data engineer at a multi-engine shop.** Runs Argo for infra jobs and
  Airflow for data. Wants one source of truth and a CI check that a pipeline is
  valid and portable before it ships. Adopts a CLI/LSP at near-zero switching cost.
- **Team evaluating or migrating orchestrators** (sharpened by the Prefect–Dagster
  consolidation). Wants to know, concretely, what a workflow loses moving between
  engines before committing.

## Painful use cases

1. A pipeline passes local review, then fails only on the Airflow scheduler hours
   later over a trivial DAG error — no pre-deploy gate caught it.
2. A team migrates from Airflow to Argo and silently loses retry-backoff and
   conditional semantics with no report of the drift.
3. Two teams maintain the "same" pipeline in two engines; the definitions diverge
   and nobody notices until production behaviour differs.

## Recommended wedge (Direction 7: workbench + compiler)

Lead with the **local-first developer workbench**; make **multi-target
compilation** a supporting feature, starting only with Argo + Airflow (which
already work). Anchor both on the canonical versioned IR, and make the novel
sliver the **capability / semantic-loss report**.

- **For it:** the only direction true in the repo *today*; demoable immediately;
  bottom-up adoption at zero switching cost; genuine whitespace (no incumbent
  owns pre-deploy engine-agnostic validation; loss diagnostics are unclaimed).
- **Serious counterargument:** "portable compiler / run-anywhere" is a graveyard
  (Couler shipped this exact thesis and still effectively targets only Argo; CWL
  stayed niche). Semantic loss makes N-target compilation leaky, and switching-pull
  is weak. The "LLVM/Terraform for workflows" analogy is unjustified: those own
  the execution substrate; a transpiler to someone else's runtime owns nothing.
- **Response:** don't sell "run anywhere." Sell **"design, validate, and de-risk
  before you commit."** The defensible value is the IR + validation + honest loss
  report, not lossless portability. Cap targets at the two that work; every new
  target is opt-in. Drop the LLVM/Terraform framing.
- **Alternative considered — ride the agent wave:** rejected. Crowded (LangGraph,
  Temporal, Kestra, Flyte), zero repo fit (needs a durable runtime FlowForge lacks),
  undemonstrable here. Let the IR represent agent steps later as that standard settles.

## Rejected directions

- **General-purpose orchestrator** — suicidal competition, no runtime, no differentiation.
- **Multi-runtime platform** — equals Kestra's/Prefect's claim; needs an engine we don't have.
- **Governance control plane (standalone)** — real enterprise pull but unsellable
  before authoring-layer adoption; keep as a later commercial layer on the IR.
- **Migration layer (as primary)** — genuine pain but services-heavy and one-time;
  demote to a feature ("import an existing DAG → IR").

## OSS / commercial split

- **OSS (bottom-up adoption):** CLI, canonical IR + JSON Schema, DAG validator,
  Argo + Airflow codegen, local runner, VS Code/LSP, capability diagnostics.
- **Commercial (team control plane):** shared/versioned IR registry, CI policy
  enforcement, cross-environment drift detection, enterprise targets (Step
  Functions, MWAA/Composer, Kestra), governance + OpenLineage overlays, SSO/audit.

## Three-year extension opportunities

Add targets (Step Functions, Kestra, Composer); represent agent/LLM steps in the
IR as that standard matures; policy-as-code governance on the IR; emit
OpenLineage from generated pipelines; "workflow diff/lint" and a shared IR/component hub.

## Confidence

**6/10.** Strong repo-fit, real pain, immediate demonstrability — but the
portability value has genuine switching-pull weakness and a prior-art graveyard.
Success hinges on winning the local-DX wedge first and *not* over-promising the compiler.

## Sources

- Prefect acquires Dagster (2026-07-13): prefect.io/prefect-acquires-dagster; Business Wire; The New Stack
- Airflow 3 GA (airflow.apache.org/blog); Airflow 3.3 release notes
- Kestra 1.0 (kestra.io/1-0)
- Union/Flyte V2; Flyte (flyte.org)
- Temporal (temporal.io); Replay 2026
- LangGraph (github.com/langchain-ai/langgraph)
- Hamilton / DagWorks (hamilton.dagworks.io)
- AWS MWAA vs Step Functions (aws.amazon.com/blogs)
- Prior art: Couler (github.com/couler-proj/couler; arXiv 2403.07608); CWL (commonwl.org)
