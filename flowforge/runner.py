"""Local in-process runner.

Executes the *real* authored task callables in topological order, passing each
task the return values of its upstream tasks (matched by parameter name), and
records real events (status, duration, captured stdout, exceptions).

Trust boundary: this runs the user's own Python in-process. It is a local
development tool and performs NO sandboxing. It must never be described as safe
execution of untrusted workflows (see docs/security/THREAT_MODEL.md).
"""

from __future__ import annotations

import io
import time
import traceback
import uuid
from contextlib import redirect_stdout
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional

from .authoring import Pipeline
from .store import Store
from .validation import validate


@dataclass
class TaskRun:
    task: str
    status: str  # succeeded | failed | skipped
    duration_ms: int
    logs: str = ""
    error: str = ""
    result_repr: str = ""


@dataclass
class RunResult:
    run_id: str
    pipeline: str
    status: str  # succeeded | failed | invalid
    tasks: List[TaskRun] = field(default_factory=list)
    validation_errors: List[str] = field(default_factory=list)

    def to_dict(self) -> dict:
        return {
            "run_id": self.run_id,
            "pipeline": self.pipeline,
            "status": self.status,
            "tasks": [t.__dict__ for t in self.tasks],
            "validation_errors": self.validation_errors,
        }


def _topo_order(pipeline: Pipeline) -> List[str]:
    names = [t.name for t in pipeline.tasks]
    edges = pipeline._infer_edges()
    indeg = {n: 0 for n in names}
    adj: Dict[str, List[str]] = {n: [] for n in names}
    for src, dst in edges:
        indeg[dst] += 1
        adj[src].append(dst)
    queue = sorted([n for n in names if indeg[n] == 0])
    order: List[str] = []
    while queue:
        n = queue.pop(0)
        order.append(n)
        for nxt in sorted(adj[n]):
            indeg[nxt] -= 1
            if indeg[nxt] == 0:
                queue.append(nxt)
        queue.sort()
    return order


def run_pipeline(pipeline: Pipeline, store: Optional[Store] = None) -> RunResult:
    """Validate then execute a pipeline locally, recording real events."""
    run_id = uuid.uuid4().hex[:12]
    spec = pipeline.to_ir()

    vr = validate(spec)
    if not vr.ok:
        result = RunResult(
            run_id=run_id,
            pipeline=pipeline.name,
            status="invalid",
            validation_errors=[d.message for d in vr.errors],
        )
        if store:
            store.start_run(run_id, pipeline.name, "local")
            store.finish_run(run_id, "invalid")
        return result

    if store:
        store.start_run(run_id, pipeline.name, "local")

    task_by_name = {t.name: t for t in pipeline.tasks}
    order = _topo_order(pipeline)
    results: Dict[str, Any] = {}
    run = RunResult(run_id=run_id, pipeline=pipeline.name, status="succeeded")
    failed = False

    for tname in order:
        tdef = task_by_name[tname]
        upstream = spec.upstream(tname)
        if failed:
            tr = TaskRun(task=tname, status="skipped", duration_ms=0, logs="upstream task failed")
            run.tasks.append(tr)
            if store:
                now = time.time()
                store.record_task(run_id, tname, "skipped", now, now, tr.logs)
            continue

        kwargs = {u: results[u] for u in upstream if u in tdef.params}
        buf = io.StringIO()
        started = time.time()
        try:
            with redirect_stdout(buf):
                value = tdef(**kwargs)
            finished = time.time()
            results[tname] = value
            tr = TaskRun(
                task=tname,
                status="succeeded",
                duration_ms=int((finished - started) * 1000),
                logs=buf.getvalue(),
                result_repr=repr(value)[:500],
            )
            if store:
                store.record_task(run_id, tname, "succeeded", started, finished, buf.getvalue())
        except Exception as exc:  # noqa: BLE001 - user code, report faithfully
            finished = time.time()
            failed = True
            err = f"{type(exc).__name__}: {exc}\n{traceback.format_exc()}"
            tr = TaskRun(
                task=tname,
                status="failed",
                duration_ms=int((finished - started) * 1000),
                logs=buf.getvalue(),
                error=err,
            )
            if store:
                store.record_task(run_id, tname, "failed", started, finished, buf.getvalue(), err)
        run.tasks.append(tr)

    run.status = "failed" if failed else "succeeded"
    if store:
        store.finish_run(run_id, run.status)
    return run
