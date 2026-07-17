"""Authoring API: define pipelines as decorated Python functions.

    from flowforge import task, pipeline

    @task(retries=3, timeout="30s")
    def extract():
        return [1, 2, 3]

    @task
    def transform(extract):          # depends on `extract` by parameter name
        return [x * 2 for x in extract]

    etl = pipeline("etl", tasks=[extract, transform], owner="data-team")

A task's parameters that match another task's name become edges. The same
structure is recovered statically (without executing the module) by
:mod:`flowforge.parser`, so ``ff compile`` never runs user code while
``ff run`` executes the real callables in topological order.
"""

from __future__ import annotations

import inspect
from typing import Any, Callable, List, Optional

from .ir import Edge, Handler, Metadata, PipelineSpec, RetryPolicy, Task, TaskPort


class TaskDef:
    """A callable wrapped by :func:`task`, carrying FlowForge metadata."""

    def __init__(
        self,
        func: Callable,
        name: Optional[str] = None,
        image: Optional[str] = None,
        retries: Optional[int] = None,
        timeout: Optional[str] = None,
        cpu: Optional[str] = None,
        memory: Optional[str] = None,
        description: Optional[str] = None,
    ) -> None:
        self.func = func
        self.name = name or func.__name__
        self.image = image
        self.retries = retries
        self.timeout = timeout
        self.cpu = cpu
        self.memory = memory
        self.description = description or (inspect.getdoc(func) or "").split("\n")[0]
        self.params: List[str] = list(inspect.signature(func).parameters)

    def __call__(self, *args: Any, **kwargs: Any) -> Any:
        return self.func(*args, **kwargs)

    def __repr__(self) -> str:  # pragma: no cover - debug aid
        return f"TaskDef({self.name!r})"


def task(func: Optional[Callable] = None, **kwargs: Any):
    """Decorator marking a function as a FlowForge task.

    Usable bare (``@task``) or called (``@task(retries=3)``).
    """

    def wrap(f: Callable) -> TaskDef:
        return TaskDef(f, **kwargs)

    if func is not None and callable(func) and not kwargs:
        return wrap(func)
    return wrap


class Pipeline:
    """An authored pipeline: metadata plus an ordered set of tasks."""

    def __init__(
        self,
        name: str,
        tasks: Optional[List[TaskDef]] = None,
        version: str = "",
        owner: str = "",
        description: str = "",
    ) -> None:
        self.name = name
        self.version = version
        self.owner = owner
        self.description = description
        self.tasks: List[TaskDef] = list(tasks or [])

    def add_task(self, t: TaskDef) -> "Pipeline":
        self.tasks.append(t)
        return self

    def _infer_edges(self) -> List[tuple]:
        names = {t.name for t in self.tasks}
        edges: List[tuple] = []
        seen = set()
        for t in self.tasks:
            for p in t.params:
                if p in names and p != t.name and (p, t.name) not in seen:
                    seen.add((p, t.name))
                    edges.append((p, t.name))
        return edges

    def to_ir(self) -> PipelineSpec:
        """Build the canonical IR from the authored pipeline."""
        edges = self._infer_edges()
        indeg = {t.name: 0 for t in self.tasks}
        outdeg = {t.name: 0 for t in self.tasks}
        for src, dst in edges:
            outdeg[src] += 1
            indeg[dst] += 1

        spec = PipelineSpec(
            metadata=Metadata(
                name=self.name,
                version=self.version,
                owner=self.owner,
                description=self.description,
            )
        )
        for t in self.tasks:
            if indeg[t.name] == 0:
                ttype = "Source"
            elif outdeg[t.name] == 0:
                ttype = "Sink"
            else:
                ttype = "Transform"

            exec_cfg: dict = {}
            if t.image:
                exec_cfg.setdefault("argo", {})["image"] = t.image
            resources = {}
            if t.cpu:
                resources["cpu"] = t.cpu
            if t.memory:
                resources["memory"] = t.memory

            spec.tasks[t.name] = Task(
                type=ttype,
                handler=Handler(type="python", source=f"{t.func.__module__}:{t.func.__name__}"),
                description=t.description,
                retry=RetryPolicy(max_attempts=t.retries) if t.retries else None,
                timeout=t.timeout,
                resources=resources,
                executor_config=exec_cfg,
            )
        for src, dst in edges:
            spec.edges.append(Edge(TaskPort(src, "out"), TaskPort(dst, "in")))
        return spec


def pipeline(
    name: str,
    tasks: Optional[List[TaskDef]] = None,
    version: str = "",
    owner: str = "",
    description: str = "",
) -> Pipeline:
    return Pipeline(name, tasks=tasks, version=version, owner=owner, description=description)
