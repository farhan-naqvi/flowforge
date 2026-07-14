"""Static (AST-based) parser: Python source -> canonical IR.

This is the *safe* front door used by ``ff compile`` and ``ff validate``:
it never imports or executes the user's module. It recognises the same
authoring model as :mod:`flowforge.authoring` (``@task`` functions assembled
by a ``pipeline(...)`` call) and produces byte-identical IR to
``Pipeline.to_ir`` so that compiling a file and importing it agree.

The in-process runner (:mod:`flowforge.runner`) uses the live objects instead,
because it needs the real callables to execute.
"""

from __future__ import annotations

import ast
from typing import Any, Dict, List, Optional

from .ir import Edge, Handler, Metadata, PipelineSpec, RetryPolicy, Task, TaskPort


class ParseError(ValueError):
    pass


def _literal(node: ast.AST) -> Optional[object]:
    try:
        return ast.literal_eval(node)
    except Exception:
        return None


def _decorator_name(dec: ast.AST) -> Optional[str]:
    if isinstance(dec, ast.Name):
        return dec.id
    if isinstance(dec, ast.Call) and isinstance(dec.func, ast.Name):
        return dec.func.id
    if isinstance(dec, ast.Attribute):
        return dec.attr
    if isinstance(dec, ast.Call) and isinstance(dec.func, ast.Attribute):
        return dec.func.attr
    return None


def _kwargs(call: ast.Call) -> Dict[str, object]:
    out: Dict[str, object] = {}
    for kw in call.keywords:
        if kw.arg is None:
            continue
        lit = _literal(kw.value)
        if lit is not None:
            out[kw.arg] = lit
    return out


class _TaskInfo:
    def __init__(self, name: str, params: List[str], opts: Dict[str, object], doc: str) -> None:
        self.name = name
        self.params = params
        self.opts = opts
        self.doc = doc


def parse_module(source: str, module_name: str = "pipeline_module") -> PipelineSpec:
    """Parse Python *source* defining a FlowForge pipeline into canonical IR."""
    try:
        tree = ast.parse(source)
    except SyntaxError as exc:  # pragma: no cover - passthrough
        raise ParseError(f"syntax error: {exc}") from exc

    tasks_by_func: Dict[str, _TaskInfo] = {}

    # 1. Collect @task-decorated functions.
    for node in ast.walk(tree):
        if not isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            continue
        opts: Dict[str, object] = {}
        is_task = False
        for dec in node.decorator_list:
            if _decorator_name(dec) == "task":
                is_task = True
                if isinstance(dec, ast.Call):
                    opts = _kwargs(dec)
        if not is_task:
            continue
        name = str(opts.get("name", node.name))
        params = [a.arg for a in node.args.args]
        doc = (ast.get_docstring(node) or "").split("\n")[0]
        tasks_by_func[node.name] = _TaskInfo(name, params, opts, doc)

    if not tasks_by_func:
        raise ParseError("no @task-decorated functions found")

    # 2. Find the pipeline(...) assembly call.
    pipe_meta: Dict[str, object] = {}
    included_funcs: List[str] = []
    for node in ast.walk(tree):
        if not isinstance(node, ast.Call):
            continue
        fname = node.func.id if isinstance(node.func, ast.Name) else (
            node.func.attr if isinstance(node.func, ast.Attribute) else None
        )
        if fname != "pipeline":
            continue
        pipe_meta = _kwargs(node)
        if node.args:
            first = _literal(node.args[0])
            if isinstance(first, str):
                pipe_meta.setdefault("name", first)
        # tasks=[...] list of Names
        for kw in node.keywords:
            if kw.arg == "tasks" and isinstance(kw.value, (ast.List, ast.Tuple)):
                for elt in kw.value.elts:
                    if isinstance(elt, ast.Name):
                        included_funcs.append(elt.id)
        break

    # If no explicit tasks list, include every discovered task (source order).
    ordered_funcs = [f for f in tasks_by_func if not included_funcs] or [
        f for f in included_funcs if f in tasks_by_func
    ]
    if not ordered_funcs:
        ordered_funcs = list(tasks_by_func)

    name = str(pipe_meta.get("name") or module_name)
    infos = [tasks_by_func[f] for f in ordered_funcs]
    return _build_ir(name, pipe_meta, infos)


def _build_ir(name: str, pipe_meta: Dict[str, object], infos: List[_TaskInfo]) -> PipelineSpec:
    task_names = {i.name for i in infos}
    edges: List[tuple] = []
    seen = set()
    for info in infos:
        for p in info.params:
            if p in task_names and p != info.name and (p, info.name) not in seen:
                seen.add((p, info.name))
                edges.append((p, info.name))

    indeg = {i.name: 0 for i in infos}
    outdeg = {i.name: 0 for i in infos}
    for src, dst in edges:
        outdeg[src] += 1
        indeg[dst] += 1

    spec = PipelineSpec(
        metadata=Metadata(
            name=name,
            version=str(pipe_meta.get("version", "")),
            owner=str(pipe_meta.get("owner", "")),
            description=str(pipe_meta.get("description", "")),
        )
    )
    for info in infos:
        if indeg[info.name] == 0:
            ttype = "Source"
        elif outdeg[info.name] == 0:
            ttype = "Sink"
        else:
            ttype = "Transform"

        opts = info.opts
        exec_cfg: Dict[str, Any] = {}
        image = opts.get("image")
        if isinstance(image, str):
            exec_cfg.setdefault("argo", {})["image"] = image
        resources: Dict[str, str] = {}
        if isinstance(opts.get("cpu"), str):
            resources["cpu"] = str(opts["cpu"])
        if isinstance(opts.get("memory"), str):
            resources["memory"] = str(opts["memory"])
        retries = opts.get("retries")
        timeout = opts.get("timeout")

        spec.tasks[info.name] = Task(
            type=ttype,
            handler=Handler(type="python", source=f"{name}:{info.name}"),
            description=info.doc,
            retry=RetryPolicy(max_attempts=int(retries)) if isinstance(retries, int) else None,
            timeout=str(timeout) if isinstance(timeout, str) else None,
            resources=resources,
            executor_config=exec_cfg,
        )
    for src, dst in edges:
        spec.edges.append(Edge(TaskPort(src, "out"), TaskPort(dst, "in")))
    return spec


def parse_file(path: str) -> PipelineSpec:
    with open(path, "r", encoding="utf-8") as fh:
        return parse_module(fh.read(), module_name=_stem(path))


def _stem(path: str) -> str:
    import os

    base = os.path.basename(path)
    return base[:-3] if base.endswith(".py") else base
