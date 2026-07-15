"""The ``covenant`` CLI: plan → verify → compile, and a one-command demo.

Every command runs the same trust core the UI and CI use.
"""

from __future__ import annotations

import argparse
import os
import sys
from typing import Optional

from . import __version__
from .compile import compile_argo
from .model import Intent
from .odcs import load_contract
from .planner import PlanError, plan_from_intent
from .verify import verify_plan

GREEN, RED, YELLOW, DIM, RESET = "\033[32m", "\033[31m", "\033[33m", "\033[2m", "\033[0m"


def _c(text, code, on):
    return f"{code}{text}{RESET}" if on else text


def _load_and_plan(intent_path: str, base_dir: str):
    intent = Intent.load(intent_path)
    return intent, plan_from_intent(intent, base_dir=base_dir)


def cmd_plan(args) -> int:
    _, plan = _load_and_plan(args.intent, args.base_dir)
    tty = sys.stdout.isatty()
    if args.output:
        with open(args.output, "w", encoding="utf-8") as fh:
            fh.write(plan.to_yaml())
        print(f"wrote {args.output}")
    else:
        sys.stdout.write(plan.to_yaml())
    if plan.conformance.ok:
        print(_c("✓ plan conforms to target contract (schema inference)", GREEN, tty))
        return 0
    print(_c("✗ plan does NOT conform to target contract:", RED, tty))
    for p in plan.conformance.problems:
        print(f"  - {p}")
    return 1


def cmd_verify(args) -> int:
    tty = sys.stdout.isatty()
    intent, plan = _load_and_plan(args.intent, args.base_dir)
    target = load_contract(os.path.join(args.base_dir, plan.target_contract))

    print(f"plan: {plan.data_product}  ({len(plan.steps)} steps)")
    if not plan.conformance.ok:
        print(_c("✗ static conformance failed (no data run):", RED, tty))
        for p in plan.conformance.problems:
            print(f"  - {p}")
        return 1
    print(_c("✓ static: schema inference conforms to target", GREEN, tty))

    verdict = verify_plan(plan, target, base_dir=args.base_dir)
    print(f"local run on synthetic data: {verdict.row_count} rows")
    if verdict.ok:
        print(_c("✓ dynamic: output satisfies target schema + quality rules", GREEN, tty))
        for row in verdict.sample[:3]:
            print(_c(f"    {row}", DIM, tty))
        return 0
    print(_c("✗ dynamic verification failed:", RED, tty))
    for p in verdict.schema_problems + verdict.quality_problems:
        print(f"  - {p}")
    return 1


def cmd_compile(args) -> int:
    tty = sys.stdout.isatty()
    _, plan = _load_and_plan(args.intent, args.base_dir)
    if not plan.conformance.ok:
        print(_c("✗ refusing to compile: plan does not conform to target contract.", RED, tty),
              file=sys.stderr)
        return 1
    artifact = compile_argo(plan, plan_path=args.plan_path or args.intent)
    if args.output:
        with open(args.output, "w", encoding="utf-8") as fh:
            fh.write(artifact)
        print(f"wrote {args.output}")
    else:
        sys.stdout.write(artifact)
    return 0


def cmd_demo(args) -> int:
    tty = sys.stdout.isatty()
    base = args.base_dir
    intent_path = os.path.join(base, "covenant-contracts/intents/sales/daily_orders/pipeline.intent.yaml")
    print(_c("Covenant demo — contract-first, verify-before-run", "\033[1m", tty))
    print(f"\n[1] Intent: {DIM}{intent_path}{RESET}" if tty else f"\n[1] Intent: {intent_path}")
    intent, plan = _load_and_plan(intent_path, base)
    print(f"[2] Planner composed {len(plan.steps)} steps from the primitive library")
    print("[3] Static schema-inference conformance vs target contract:")
    print("   ", _c("✓ conforms", GREEN, tty) if plan.conformance.ok
          else _c("✗ " + "; ".join(plan.conformance.problems), RED, tty))
    target = load_contract(os.path.join(base, plan.target_contract))
    verdict = verify_plan(plan, target, base_dir=base)
    print(f"[4] Local run on synthetic + edge-case data: {verdict.row_count} rows")
    print("   ", _c("✓ output satisfies target schema + quality rules", GREEN, tty) if verdict.ok
          else _c("✗ " + "; ".join(verdict.schema_problems + verdict.quality_problems), RED, tty))
    for row in verdict.sample[:3]:
        print(_c(f"      {row}", DIM, tty))
    if verdict.ok:
        art = compile_argo(plan, plan_path="plans/sales/daily_orders/pipeline.plan.yaml")
        print(f"[5] Compiled Argo/Spark artifact ({len(art.splitlines())} lines) — ready for one-click run")
    print(_c("\nGreen check = trusted to run at full scale.", "\033[1m", tty))
    return 0 if verdict.ok else 1


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog="covenant", description="Covenant contract-first pipeline workbench")
    p.add_argument("--version", action="version", version=f"covenant {__version__}")
    p.add_argument("--base-dir", default=".", help="repo root for resolving contract paths")
    sub = p.add_subparsers(dest="command", required=True)

    pl = sub.add_parser("plan", help="compose a plan from an intent + show static conformance")
    pl.add_argument("intent")
    pl.add_argument("--output", "-o")
    pl.set_defaults(func=cmd_plan)

    ve = sub.add_parser("verify", help="static + local dynamic verification against the target contract")
    ve.add_argument("intent")
    ve.set_defaults(func=cmd_verify)

    co = sub.add_parser("compile", help="compile a conformant plan to an Argo/Spark artifact")
    co.add_argument("intent")
    co.add_argument("--plan-path", help="committed plan path to reference in the artifact")
    co.add_argument("--output", "-o")
    co.set_defaults(func=cmd_compile)

    de = sub.add_parser("demo", help="run the sales/daily_orders example end-to-end")
    de.set_defaults(func=cmd_demo)

    sv = sub.add_parser("serve", help="serve the Covenant workbench UI + JSON API")
    sv.add_argument("--host", default="127.0.0.1")
    sv.add_argument("--port", type=int, default=8090)
    sv.set_defaults(func=cmd_serve)
    return p


def cmd_serve(args) -> int:
    from .server import serve

    serve(host=args.host, port=args.port, base_dir=args.base_dir)
    return 0


def main(argv: Optional[list] = None) -> int:
    args = build_parser().parse_args(argv)
    # base-dir is a top-level arg; subparsers inherit via the same namespace.
    if not hasattr(args, "base_dir"):
        args.base_dir = "."
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
