#!/usr/bin/env python3
"""One-command guided demo of the real FlowForge path.

    python scripts/demo.py

Shows: author -> canonical IR -> validate -> capability report -> compile
(Argo + Airflow) -> real local run -> and a FAILURE case (a cyclic pipeline
that validation rejects and refuses to compile). No fabricated data.
"""

from __future__ import annotations

import os
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, ROOT)

BOLD, DIM, RESET = "\033[1m", "\033[2m", "\033[0m"


def step(n: int, title: str) -> None:
    print(f"\n{BOLD}[{n}] {title}{RESET}")


def run(cmd: list) -> int:
    print(f"{DIM}$ {' '.join(cmd)}{RESET}")
    return subprocess.call(cmd, cwd=ROOT)


def main() -> int:
    py = sys.executable
    ff = [py, "-m", "flowforge.cli"]
    simple = "examples/simple_etl.py"
    broken = "examples/broken_cycle.py"
    env = dict(os.environ, PYTHONPATH=ROOT, PYTHONDONTWRITEBYTECODE="1")
    os.environ.update(env)

    print(f"{BOLD}FlowForge demo — the real end-to-end path{RESET}")

    step(1, "Author: a pipeline is plain decorated Python")
    run(["sed", "-n", "1,40p", os.path.join(ROOT, simple)]) if os.name != "nt" else None

    step(2, "Canonical IR (flowforge.io/v1) parsed statically, without running your code")
    run(ff + ["ir", simple])

    step(3, "Validate: DAG, schema, and injection-safe identifiers")
    run(ff + ["validate", simple])

    step(4, "Capability report: what each target preserves vs drops")
    run(ff + ["capability", simple])

    step(5, "Compile to Argo Workflows (safe YAML emission)")
    run(ff + ["compile", simple, "argo"])

    step(6, "Compile to Apache Airflow (Airflow 3 style, safe emission)")
    run(ff + ["compile", simple, "airflow"])

    step(7, "Run locally: real in-process execution with real logs")
    run(ff + ["run", simple])

    step(8, "FAILURE case: a cyclic pipeline is rejected, not silently 'fixed'")
    rc = run(ff + ["validate", broken])
    print(f"{DIM}(validate exited {rc}; compile would refuse){RESET}")
    run(ff + ["compile", broken, "argo"])

    print(f"\n{BOLD}Done.{RESET} Try the UI:  {DIM}ff serve --workspace examples{RESET}  then open http://127.0.0.1:8080")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
