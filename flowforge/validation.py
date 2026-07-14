"""Validation: graph correctness, schema conformance, and injection safety.

Diagnostics are structured so the CLI, API and UI render the same results.
Identifier validation (ADR-005) is a security control, not a nicety: unchecked
names flow into generated Airflow Python and Argo YAML, so an unsafe pipeline
name is a code-injection vector. We reject unsafe identifiers *and* the
compilers emit through safe serializers (defense in depth).
"""

from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Dict, List

from .ir import PipelineSpec

# DNS-1123-ish: start with a letter, then letters/digits/-/_ up to 63 chars.
IDENT_RE = re.compile(r"^[A-Za-z][A-Za-z0-9_-]{0,62}$")
OWNER_RE = re.compile(r"^[A-Za-z0-9_.\- ]{0,64}$")
IMAGE_RE = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._/:@-]{0,254}$")

SEVERITY_ERROR = "error"
SEVERITY_WARNING = "warning"


@dataclass
class Diagnostic:
    severity: str
    code: str
    message: str
    location: str = ""

    def to_dict(self) -> dict:
        return {
            "severity": self.severity,
            "code": self.code,
            "message": self.message,
            "location": self.location,
        }


@dataclass
class ValidationResult:
    diagnostics: List[Diagnostic]

    @property
    def errors(self) -> List[Diagnostic]:
        return [d for d in self.diagnostics if d.severity == SEVERITY_ERROR]

    @property
    def warnings(self) -> List[Diagnostic]:
        return [d for d in self.diagnostics if d.severity == SEVERITY_WARNING]

    @property
    def ok(self) -> bool:
        return not self.errors

    def to_dict(self) -> dict:
        return {
            "ok": self.ok,
            "errors": [d.to_dict() for d in self.errors],
            "warnings": [d.to_dict() for d in self.warnings],
        }


def validate(spec: PipelineSpec) -> ValidationResult:
    diags: List[Diagnostic] = []
    _check_identifiers(spec, diags)
    _check_structure(spec, diags)
    _check_graph(spec, diags)
    return ValidationResult(diags)


def _err(diags, code, message, location=""):
    diags.append(Diagnostic(SEVERITY_ERROR, code, message, location))


def _warn(diags, code, message, location=""):
    diags.append(Diagnostic(SEVERITY_WARNING, code, message, location))


def _check_identifiers(spec: PipelineSpec, diags: List[Diagnostic]) -> None:
    name = spec.metadata.name
    if not name:
        _err(diags, "name.required", "pipeline name is required", "metadata.name")
    elif not IDENT_RE.match(name):
        _err(
            diags,
            "name.unsafe",
            f"pipeline name {name!r} must match {IDENT_RE.pattern} "
            "(prevents code injection into generated artifacts)",
            "metadata.name",
        )
    if spec.metadata.owner and not OWNER_RE.match(spec.metadata.owner):
        _err(diags, "owner.unsafe", f"owner {spec.metadata.owner!r} contains unsafe characters", "metadata.owner")

    for tid, task in spec.tasks.items():
        if not IDENT_RE.match(tid):
            _err(
                diags,
                "task.id.unsafe",
                f"task id {tid!r} must match {IDENT_RE.pattern} "
                "(prevents code injection into generated artifacts)",
                f"tasks.{tid}",
            )
        image = task.executor_config.get("argo", {}).get("image") if task.executor_config else None
        if image and not IMAGE_RE.match(str(image)):
            _err(diags, "task.image.unsafe", f"task {tid!r} image {image!r} is not a valid container reference", f"tasks.{tid}")
        if task.timeout and not re.match(r"^\d+[smhd]?$", task.timeout):
            _warn(diags, "task.timeout.format", f"task {tid!r} timeout {task.timeout!r} is not a recognised duration (e.g. '30s', '1h')", f"tasks.{tid}")
        if task.retry and task.retry.max_attempts < 0:
            _err(diags, "task.retry.negative", f"task {tid!r} has negative retry maxAttempts", f"tasks.{tid}")


def _check_structure(spec: PipelineSpec, diags: List[Diagnostic]) -> None:
    if spec.api_version != "flowforge.io/v1":
        _err(diags, "apiVersion.invalid", f"unsupported apiVersion {spec.api_version!r}", "apiVersion")
    if not spec.tasks:
        _err(diags, "tasks.empty", "pipeline must define at least one task", "tasks")
    for tid, task in spec.tasks.items():
        if task.type not in ("Source", "Transform", "Sink", "Conditional", "Schedule"):
            _err(diags, "task.type.invalid", f"task {tid!r} has invalid type {task.type!r}", f"tasks.{tid}")


def _check_graph(spec: PipelineSpec, diags: List[Diagnostic]) -> None:
    ids = set(spec.tasks)
    for i, edge in enumerate(spec.edges):
        if edge.from_.task not in ids:
            _err(diags, "edge.source.missing", f"edge {i} references unknown source task {edge.from_.task!r}", f"edges[{i}]")
        if edge.to.task not in ids:
            _err(diags, "edge.target.missing", f"edge {i} references unknown target task {edge.to.task!r}", f"edges[{i}]")

    # Cycle detection (DFS colouring).
    adj: Dict[str, List[str]] = {tid: [] for tid in ids}
    for edge in spec.edges:
        if edge.from_.task in adj and edge.to.task in ids:
            adj[edge.from_.task].append(edge.to.task)
    color: Dict[str, int] = {tid: 0 for tid in ids}
    cycle_found = [False]

    def visit(node: str) -> None:
        color[node] = 1
        for nxt in adj.get(node, []):
            if color[nxt] == 1:
                cycle_found[0] = True
            elif color[nxt] == 0:
                visit(nxt)
        color[node] = 2

    for tid in ids:
        if color[tid] == 0:
            visit(tid)
    if cycle_found[0]:
        _err(diags, "graph.cycle", "pipeline contains a cycle; workflows must be acyclic", "edges")

    # Reachability: warn about isolated multi-task pipelines.
    if len(ids) > 1:
        connected = set()
        for edge in spec.edges:
            connected.add(edge.from_.task)
            connected.add(edge.to.task)
        for tid in ids - connected:
            _warn(diags, "task.isolated", f"task {tid!r} has no edges (isolated node)", f"tasks.{tid}")
