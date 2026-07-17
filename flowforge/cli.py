"""The ``ff`` command line: validate, inspect IR, capability report, compile,
run locally, and serve the workbench UI. Every subcommand calls the same
library functions the API and UI use."""

from __future__ import annotations

import argparse
import sys
from typing import List, Optional

from . import __version__
from .capability import LOSSY, UNSUPPORTED, analyze
from .compilers import TARGETS, compile as compile_target
from .parser import ParseError, parse_file
from .validation import validate

GREEN, RED, YELLOW, DIM, RESET = "\033[32m", "\033[31m", "\033[33m", "\033[2m", "\033[0m"


def _color(text: str, code: str, use: bool) -> str:
    return f"{code}{text}{RESET}" if use else text


def _load_spec(path: str):
    try:
        return parse_file(path)
    except (ParseError, FileNotFoundError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        raise SystemExit(2)


def cmd_validate(args) -> int:
    spec = _load_spec(args.file)
    result = validate(spec)
    tty = sys.stdout.isatty()
    for d in result.errors:
        print(_color("error", RED, tty) + f"  [{d.code}] {d.message}" + (f"  ({d.location})" if d.location else ""))
    for d in result.warnings:
        print(_color("warn", YELLOW, tty) + f"   [{d.code}] {d.message}" + (f"  ({d.location})" if d.location else ""))
    if result.ok:
        print(_color("✓", GREEN, tty) + f" {spec.metadata.name}: valid "
              f"({len(spec.tasks)} tasks, {len(spec.edges)} edges, {len(result.warnings)} warnings)")
        return 0
    print(_color("✗", RED, tty) + f" {spec.metadata.name}: {len(result.errors)} error(s)")
    return 1


def cmd_ir(args) -> int:
    spec = _load_spec(args.file)
    print(spec.to_json())
    return 0


def cmd_capability(args) -> int:
    spec = _load_spec(args.file)
    tty = sys.stdout.isatty()
    targets = [args.target] if args.target else list(TARGETS)
    for target in targets:
        report = analyze(spec, target)
        print(f"\n{target}:")
        for item in report.items:
            if not item.used:
                mark, code = "·", DIM
            elif item.level == UNSUPPORTED:
                mark, code = "✗", RED
            elif item.level == LOSSY:
                mark, code = "~", YELLOW
            else:
                mark, code = "✓", GREEN
            used = "" if item.used else _color(" (not used)", DIM, tty)
            print(f"  {_color(mark, code, tty)} {item.feature:16} {item.level:12} {item.note}{used}")
        s = report.to_dict()["summary"]
        print(f"  → {s['lossy']} lossy, {s['unsupported']} unsupported (of features actually used)")
    return 0


def cmd_compile(args) -> int:
    spec = _load_spec(args.file)
    result = validate(spec)
    if not result.ok:
        print(f"✗ refusing to compile {spec.metadata.name}: {len(result.errors)} validation error(s). "
              f"Run 'ff validate {args.file}'.", file=sys.stderr)
        return 1
    artifact = compile_target(spec, args.target)
    if args.output:
        with open(args.output, "w", encoding="utf-8") as fh:
            fh.write(artifact)
        print(f"wrote {args.output}")
    else:
        sys.stdout.write(artifact)
    report = analyze(spec, args.target)
    if report.lossy_used or report.unsupported_used:
        print(f"\n# capability notes for {args.target}:", file=sys.stderr)
        for item in report.lossy_used + report.unsupported_used:
            print(f"#   {item.level}: {item.feature} — {item.note}", file=sys.stderr)
    return 0


def cmd_run(args) -> int:
    from .loader import LoadError, load_pipeline
    from .runner import run_pipeline
    from .store import Store

    try:
        pipeline = load_pipeline(args.file, name=args.name)
    except LoadError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2
    store = Store(args.db) if args.db else Store()
    result = run_pipeline(pipeline, store=store)
    tty = sys.stdout.isatty()
    print(f"run {result.run_id}: {result.pipeline}")
    if result.status == "invalid":
        for e in result.validation_errors:
            print(_color("  error ", RED, tty) + e)
        print(_color("✗ invalid — not executed", RED, tty))
        return 1
    for t in result.tasks:
        mark = {"succeeded": (GREEN, "✓"), "failed": (RED, "✗"), "skipped": (YELLOW, "○")}[t.status]
        line = f"  {_color(mark[1], mark[0], tty)} {t.task:16} {t.status:10} {t.duration_ms:>6} ms"
        print(line)
        if t.logs.strip():
            for ln in t.logs.strip().splitlines():
                print(_color(f"      {ln}", DIM, tty))
        if t.error:
            print(_color(f"      {t.error.splitlines()[0]}", RED, tty))
    status_color = GREEN if result.status == "succeeded" else RED
    print(_color(f"{result.status}", status_color, tty) + f"  (stored in {store.path})")
    store.close()
    return 0 if result.status == "succeeded" else 1


def cmd_serve(args) -> int:
    from .server import serve

    serve(host=args.host, port=args.port, db=args.db, workspace=args.workspace)
    return 0


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog="ff", description="FlowForge portable workflow workbench")
    p.add_argument("--version", action="version", version=f"flowforge {__version__}")
    sub = p.add_subparsers(dest="command", required=True)

    v = sub.add_parser("validate", help="validate a pipeline (graph, schema, injection-safe identifiers)")
    v.add_argument("file")
    v.set_defaults(func=cmd_validate)

    i = sub.add_parser("ir", help="print the canonical IR (JSON)")
    i.add_argument("file")
    i.set_defaults(func=cmd_ir)

    c = sub.add_parser("capability", help="show what each target preserves/drops")
    c.add_argument("file")
    c.add_argument("--target", choices=list(TARGETS))
    c.set_defaults(func=cmd_capability)

    co = sub.add_parser("compile", help="compile to an orchestrator artifact")
    co.add_argument("file")
    co.add_argument("target", choices=list(TARGETS))
    co.add_argument("--output", "-o")
    co.set_defaults(func=cmd_compile)

    r = sub.add_parser("run", help="run the pipeline locally (executes your code)")
    r.add_argument("file")
    r.add_argument("--name", help="pipeline variable name if the module defines several")
    r.add_argument("--db", help="event store path (default ~/.flowforge/runs.db)")
    r.set_defaults(func=cmd_run)

    s = sub.add_parser("serve", help="serve the workbench UI + JSON API")
    s.add_argument("--host", default="127.0.0.1")
    s.add_argument("--port", type=int, default=8080)
    s.add_argument("--db")
    s.add_argument("--workspace", default=".", help="directory scanned for *.py pipelines")
    s.set_defaults(func=cmd_serve)

    return p


def main(argv: Optional[List[str]] = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
