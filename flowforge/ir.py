"""Canonical FlowForge Intermediate Representation (``flowforge.io/v1``).

This module is the single in-memory representation every FlowForge surface
(CLI, HTTP API, UI, runner, compilers) consumes. The wire contract is the JSON
Schema at ``ir/spec.json``; these dataclasses implement it in Python.

The IR is deliberately small. Fields that a compile target cannot faithfully
represent are reported by :mod:`flowforge.capability`, never dropped silently.
"""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional

API_VERSION = "flowforge.io/v1"
KIND = "Pipeline"

# Task types (mirrors ir/spec.json enum).
TASK_TYPES = ("Source", "Transform", "Sink", "Conditional", "Schedule")


@dataclass
class RetryPolicy:
    max_attempts: int
    backoff: str = "exponential"  # linear | exponential | fixed
    backoff_multiplier: Optional[float] = None
    initial_delay_seconds: Optional[int] = None

    def to_dict(self) -> Dict[str, Any]:
        out: Dict[str, Any] = {"maxAttempts": self.max_attempts, "backoff": self.backoff}
        if self.backoff_multiplier is not None:
            out["backoffMultiplier"] = self.backoff_multiplier
        if self.initial_delay_seconds is not None:
            out["initialDelaySeconds"] = self.initial_delay_seconds
        return out

    @classmethod
    def from_dict(cls, d: Dict[str, Any]) -> "RetryPolicy":
        return cls(
            max_attempts=int(d["maxAttempts"]),
            backoff=d.get("backoff", "exponential"),
            backoff_multiplier=d.get("backoffMultiplier"),
            initial_delay_seconds=d.get("initialDelaySeconds"),
        )


@dataclass
class Handler:
    type: str = "python"  # python | sql | docker | http
    source: str = ""
    env: Dict[str, str] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        out: Dict[str, Any] = {"type": self.type, "source": self.source}
        if self.env:
            out["env"] = dict(self.env)
        return out

    @classmethod
    def from_dict(cls, d: Dict[str, Any]) -> "Handler":
        return cls(type=d.get("type", "python"), source=d.get("source", ""), env=dict(d.get("env", {})))


@dataclass
class Task:
    type: str
    handler: Handler
    description: str = ""
    inputs: Dict[str, dict] = field(default_factory=dict)
    outputs: Dict[str, dict] = field(default_factory=dict)
    executor_config: Dict[str, Any] = field(default_factory=dict)
    retry: Optional[RetryPolicy] = None
    timeout: Optional[str] = None
    resources: Dict[str, str] = field(default_factory=dict)  # {"cpu": "1", "memory": "2Gi"}
    metadata: Dict[str, str] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        out: Dict[str, Any] = {"type": self.type, "handler": self.handler.to_dict()}
        if self.description:
            out["description"] = self.description
        if self.inputs:
            out["inputs"] = self.inputs
        if self.outputs:
            out["outputs"] = self.outputs
        if self.executor_config:
            out["executorConfig"] = self.executor_config
        if self.retry is not None:
            out["retry"] = self.retry.to_dict()
        if self.timeout:
            out["timeout"] = self.timeout
        if self.resources:
            out["resources"] = self.resources
        if self.metadata:
            out["metadata"] = self.metadata
        return out

    @classmethod
    def from_dict(cls, d: Dict[str, Any]) -> "Task":
        return cls(
            type=d["type"],
            handler=Handler.from_dict(d.get("handler", {})),
            description=d.get("description", ""),
            inputs=d.get("inputs", {}) or {},
            outputs=d.get("outputs", {}) or {},
            executor_config=d.get("executorConfig", {}) or {},
            retry=RetryPolicy.from_dict(d["retry"]) if d.get("retry") else None,
            timeout=d.get("timeout"),
            resources=d.get("resources", {}) or {},
            metadata=d.get("metadata", {}) or {},
        )


@dataclass
class TaskPort:
    task: str
    port: str

    def to_dict(self) -> Dict[str, str]:
        return {"task": self.task, "port": self.port}


@dataclass
class Edge:
    from_: TaskPort
    to: TaskPort

    def to_dict(self) -> Dict[str, Any]:
        return {"from": self.from_.to_dict(), "to": self.to.to_dict()}

    @classmethod
    def from_dict(cls, d: Dict[str, Any]) -> "Edge":
        f, t = d["from"], d["to"]
        return cls(TaskPort(f["task"], f.get("port", "out")), TaskPort(t["task"], t.get("port", "in")))


@dataclass
class Metadata:
    name: str
    version: str = ""
    owner: str = ""
    description: str = ""
    namespace: str = ""
    tags: Dict[str, str] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        out: Dict[str, Any] = {"name": self.name}
        for k in ("version", "owner", "description", "namespace"):
            v = getattr(self, k)
            if v:
                out[k] = v
        if self.tags:
            out["tags"] = self.tags
        return out

    @classmethod
    def from_dict(cls, d: Dict[str, Any]) -> "Metadata":
        return cls(
            name=d.get("name", ""),
            version=d.get("version", ""),
            owner=d.get("owner", ""),
            description=d.get("description", ""),
            namespace=d.get("namespace", ""),
            tags=dict(d.get("tags", {})),
        )


@dataclass
class PipelineSpec:
    metadata: Metadata
    tasks: Dict[str, Task] = field(default_factory=dict)
    edges: List[Edge] = field(default_factory=list)
    api_version: str = API_VERSION
    kind: str = KIND

    def to_dict(self) -> Dict[str, Any]:
        return {
            "apiVersion": self.api_version,
            "kind": self.kind,
            "metadata": self.metadata.to_dict(),
            "tasks": {tid: t.to_dict() for tid, t in self.tasks.items()},
            "edges": [e.to_dict() for e in self.edges],
        }

    def to_json(self, indent: int = 2) -> str:
        return json.dumps(self.to_dict(), indent=indent)

    @classmethod
    def from_dict(cls, d: Dict[str, Any]) -> "PipelineSpec":
        tasks = {tid: Task.from_dict(t) for tid, t in (d.get("tasks") or {}).items()}
        edges = [Edge.from_dict(e) for e in (d.get("edges") or [])]
        return cls(
            metadata=Metadata.from_dict(d.get("metadata", {})),
            tasks=tasks,
            edges=edges,
            api_version=d.get("apiVersion", API_VERSION),
            kind=d.get("kind", KIND),
        )

    @classmethod
    def from_json(cls, text: str) -> "PipelineSpec":
        return cls.from_dict(json.loads(text))

    # -- convenience -----------------------------------------------------

    def upstream(self, task_id: str) -> List[str]:
        return [e.from_.task for e in self.edges if e.to.task == task_id]

    def downstream(self, task_id: str) -> List[str]:
        return [e.to.task for e in self.edges if e.from_.task == task_id]
