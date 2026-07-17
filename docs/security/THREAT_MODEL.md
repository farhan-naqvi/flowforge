# FlowForge Threat Model & Security Baseline

Scope: the FlowForge MVP — a **local-first developer tool**: a CLI and a
loopback web server that parse Python pipeline files, compile them to Argo/Airflow
artifacts, and (on request) execute the user's own pipeline code locally, storing
run events in a local SQLite file.

This document states what FlowForge defends against, what it deliberately does
not, and what it must never claim. It reflects the code as shipped on
`claude/flowforge-core-hardening`.

## 1. Assets & trust boundaries

| Asset | Notes |
|---|---|
| The developer's workstation | primary thing to protect |
| Pipeline source files (`*.py`) | may be first-party (trusted) or third-party (untrusted) |
| Generated Argo/Airflow artifacts | deployed elsewhere; must not carry injected code |
| Local run/event store (SQLite) | run history and captured logs |

Trust boundaries:

- **T1 — the user's own pipeline code.** Running it locally is the point of the
  tool. Trusted by definition.
- **T2 — third-party pipeline files** a developer clones or downloads. Parsing
  and compiling these is safe (AST only, no execution); **running** them is not
  and requires the same trust as running any downloaded script.
- **T3 — network.** The CLI has no network surface. The workbench server is a
  local HTTP surface reachable, if misconfigured, by other hosts or by web pages
  in the developer's browser.

## 2. What FlowForge defends against (implemented)

| Threat | Defense | Where |
|---|---|---|
| **Code injection into generated Airflow/Argo** via pipeline/task/owner names (proven RCE in the pre-hardening code) | Identifier allowlist `^[A-Za-z][A-Za-z0-9_-]{0,62}$` enforced at validation; **compile refuses** on validation errors; Argo emitted via `yaml.safe_dump`, Airflow strings via `repr()` | `flowforge/validation.py`, `flowforge/compilers/*` |
| **Compiling from untrusted files** executing code | `ff compile`/`ff validate` and the server use the **AST parser only** — they never import user modules | `flowforge/parser.py` |
| **DNS-rebinding / cross-origin writes** against the local server | `Host` header must be loopback; else 403 | `flowforge/server.py` |
| **Accidental network exposure** | server binds `127.0.0.1` by default; non-loopback requires explicit `--host` and prints a warning | `flowforge/server.py` |
| **Path traversal** to read/run files outside the workspace | resolved paths must stay within the workspace root | `flowforge/server.py` |
| **Log-injection → stored XSS** in the UI | all pipeline/log/run data rendered via `textContent`, never `innerHTML`; `Content-Security-Policy` + `X-Content-Type-Options` headers; no external/CDN assets | `flowforge/webui/app.js`, `flowforge/server.py` |
| **Unbounded payloads / DB growth** | request body cap (256 KB); per-log field cap (64 KB) | `flowforge/server.py`, `flowforge/store.py` |
| **Committed secrets / data** | `observability.db` removed from git; `*.db`/`*.sqlite` gitignored; `.env` gitignored with `.env.example` template | `.gitignore`, `.env.example` |

## 3. Out of scope for the MVP (by design)

Because this is a single-user local tool, the following are **not** provided and
must not be assumed:

- Authentication / authorization on the local API (there are no users).
- Multi-tenant isolation.
- **Sandboxed execution.** `ff run` and `POST /api/run` execute the user's own
  Python **in-process with no isolation**. This is acceptable only under trust
  boundary T1 (your own code on your own machine).
- Protection against a malicious *first-party* developer.
- TLS (loopback only).

## 4. Residual risks

- **Running downloaded pipelines (T2) is arbitrary code execution.** Treat a
  third-party `*.py` exactly like any script you would run: read it first.
- Generated artifacts embed the user's handler source by design; that source is
  the user's own code, not attacker-controlled once identifiers are validated.
- The server is multithreaded; the SQLite store is guarded by a process-wide
  lock (adequate for local single-user use, not for concurrent multi-writer
  deployments).

## 5. Statements FlowForge must never make

- "Secure execution of untrusted workflows" / "sandboxed" — it is not.
- "Enterprise-ready" / "production-grade" / "multi-tenant" / "hardened".
- Any "% test coverage" figure not measured in CI.
- "Authenticated" / "access-controlled" APIs — the local API has no auth.

## 6. Production hardening roadmap (future, honest)

Before FlowForge could run in a shared/hosted setting it would need: authn/authz
on all HTTP surfaces; sandboxed execution (containers + seccomp/gVisor, resource
quotas) for any non-first-party code; per-tenant isolation of runs and storage;
TLS; audit logging; artifact/image signing (cosign) + SBOM; and a real secrets
backend. None of these are implemented today.

## 7. Supply chain

- Runtime dependency is `PyYAML` only (used precisely to avoid hand-rolled,
  injectable YAML emission). No CDN or external assets in the UI.
- CI runs `pip-audit` (Python) and `govulncheck` (Go `ir/`); see
  [.github/workflows/ci.yml](../../.github/workflows/ci.yml).
