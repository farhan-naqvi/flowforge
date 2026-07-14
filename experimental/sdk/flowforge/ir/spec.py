"""FlowForge IR Python module - dataclasses and builders for pipeline specifications."""

from __future__ import annotations
from dataclasses import dataclass, field, asdict
from typing import Dict, List, Optional, Any
from enum import Enum
import json


class TaskType(str, Enum):
    """Task type enumeration."""
    SOURCE = "Source"
    TRANSFORM = "Transform"
    SINK = "Sink"
    CONDITIONAL = "Conditional"
    RETRY = "Retry"
    SCHEDULE = "Schedule"


@dataclass
class PipelineMetadata:
    """Pipeline metadata."""
    name: str
    version: Optional[str] = None
    namespace: Optional[str] = None
    owner: Optional[str] = None
    description: Optional[str] = None
    tags: Dict[str, str] = field(default_factory=dict)


@dataclass
class Handler:
    """Handler specification."""
    type: str  # python, sql, spark, docker, http
    source: str
    env: Dict[str, str] = field(default_factory=dict)


@dataclass
class RetryPolicy:
    """Retry configuration."""
    max_attempts: int
    backoff: str = "fixed"  # linear, exponential, fixed
    backoff_multiplier: float = 1.0
    initial_delay_seconds: int = 1


@dataclass
class CostDimension:
    """Single cost dimension."""
    unit: str
    quantity: float


@dataclass
class CostEstimate:
    """Cost estimation."""
    compute: Optional[CostDimension] = None
    storage: Optional[CostDimension] = None
    network: Optional[CostDimension] = None


@dataclass
class Task:
    """Task specification."""
    type: TaskType
    handler: Handler
    description: Optional[str] = None
    inputs: Dict[str, Dict[str, Any]] = field(default_factory=dict)
    outputs: Dict[str, Dict[str, Any]] = field(default_factory=dict)
    executor_config: Dict[str, Dict[str, Any]] = field(default_factory=dict)
    cost_estimate: Optional[CostEstimate] = None
    retry: Optional[RetryPolicy] = None
    timeout: Optional[str] = None
    metadata: Dict[str, str] = field(default_factory=dict)


@dataclass
class TaskPort:
    """Task port reference."""
    task: str
    port: str


@dataclass
class Edge:
    """Data flow edge."""
    from_task: TaskPort
    to_task: TaskPort


@dataclass
class PipelineSpec:
    """Complete pipeline specification."""
    api_version: str = "flowforge.io/v1"
    kind: str = "Pipeline"
    metadata: Optional[PipelineMetadata] = None
    tasks: Dict[str, Task] = field(default_factory=dict)
    edges: List[Edge] = field(default_factory=list)

    def validate(self) -> None:
        """Validate the pipeline specification."""
        if self.api_version != "flowforge.io/v1":
            raise ValueError(f"Invalid apiVersion: {self.api_version}")
        if self.kind != "Pipeline":
            raise ValueError(f"Invalid kind: {self.kind}")
        if not self.metadata or not self.metadata.name:
            raise ValueError("metadata.name is required")
        if not self.tasks:
            raise ValueError("At least one task is required")

        # Validate edges reference valid tasks and ports
        for i, edge in enumerate(self.edges):
            if edge.from_task.task not in self.tasks:
                raise ValueError(f"Edge {i}: source task '{edge.from_task.task}' not found")
            if edge.to_task.task not in self.tasks:
                raise ValueError(f"Edge {i}: target task '{edge.to_task.task}' not found")

            from_task = self.tasks[edge.from_task.task]
            to_task = self.tasks[edge.to_task.task]

            if edge.from_task.port not in from_task.outputs:
                raise ValueError(
                    f"Edge {i}: output port '{edge.from_task.port}' "
                    f"not found on task '{edge.from_task.task}'"
                )
            if edge.to_task.port not in to_task.inputs:
                raise ValueError(
                    f"Edge {i}: input port '{edge.to_task.port}' "
                    f"not found on task '{edge.to_task.task}'"
                )

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "apiVersion": self.api_version,
            "kind": self.kind,
            "metadata": asdict(self.metadata) if self.metadata else {},
            "tasks": {
                task_id: self._task_to_dict(task)
                for task_id, task in self.tasks.items()
            },
            "edges": [
                {
                    "from": {"task": edge.from_task.task, "port": edge.from_task.port},
                    "to": {"task": edge.to_task.task, "port": edge.to_task.port},
                }
                for edge in self.edges
            ],
        }

    def to_json(self) -> str:
        """Serialize to JSON."""
        return json.dumps(self.to_dict(), indent=2)

    @staticmethod
    def _task_to_dict(task: Task) -> Dict[str, Any]:
        """Convert task to dictionary."""
        result = {
            "type": task.type.value,
            "handler": {"type": task.handler.type, "source": task.handler.source},
        }
        if task.handler.env:
            result["handler"]["env"] = task.handler.env
        if task.description:
            result["description"] = task.description
        if task.inputs:
            result["inputs"] = task.inputs
        if task.outputs:
            result["outputs"] = task.outputs
        if task.executor_config:
            result["executorConfig"] = task.executor_config
        if task.cost_estimate:
            result["costEstimate"] = asdict(task.cost_estimate)
        if task.retry:
            result["retry"] = {
                "maxAttempts": task.retry.max_attempts,
                "backoff": task.retry.backoff,
                "backoffMultiplier": task.retry.backoff_multiplier,
                "initialDelaySeconds": task.retry.initial_delay_seconds,
            }
        if task.timeout:
            result["timeout"] = task.timeout
        if task.metadata:
            result["metadata"] = task.metadata
        return result

    @staticmethod
    def from_json(json_str: str) -> PipelineSpec:
        """Deserialize from JSON."""
        data = json.loads(json_str)
        return PipelineSpec.from_dict(data)

    @staticmethod
    def from_dict(data: Dict[str, Any]) -> PipelineSpec:
        """Create from dictionary."""
        metadata_data = data.get("metadata", {})
        metadata = PipelineMetadata(
            name=metadata_data.get("name"),
            version=metadata_data.get("version"),
            namespace=metadata_data.get("namespace"),
            owner=metadata_data.get("owner"),
            description=metadata_data.get("description"),
            tags=metadata_data.get("tags", {}),
        )

        tasks = {}
        for task_id, task_data in data.get("tasks", {}).items():
            handler_data = task_data.get("handler", {})
            handler = Handler(
                type=handler_data.get("type"),
                source=handler_data.get("source"),
                env=handler_data.get("env", {}),
            )

            retry_data = task_data.get("retry")
            retry = None
            if retry_data:
                retry = RetryPolicy(
                    max_attempts=retry_data.get("maxAttempts", 1),
                    backoff=retry_data.get("backoff", "fixed"),
                    backoff_multiplier=retry_data.get("backoffMultiplier", 1.0),
                    initial_delay_seconds=retry_data.get("initialDelaySeconds", 1),
                )

            cost_data = task_data.get("costEstimate")
            cost_estimate = None
            if cost_data:
                cost_estimate = CostEstimate(
                    compute=CostDimension(**cost_data["compute"]) if cost_data.get("compute") else None,
                    storage=CostDimension(**cost_data["storage"]) if cost_data.get("storage") else None,
                    network=CostDimension(**cost_data["network"]) if cost_data.get("network") else None,
                )

            task = Task(
                type=TaskType(task_data.get("type")),
                handler=handler,
                description=task_data.get("description"),
                inputs=task_data.get("inputs", {}),
                outputs=task_data.get("outputs", {}),
                executor_config=task_data.get("executorConfig", {}),
                cost_estimate=cost_estimate,
                retry=retry,
                timeout=task_data.get("timeout"),
                metadata=task_data.get("metadata", {}),
            )
            tasks[task_id] = task

        edges = []
        for edge_data in data.get("edges", []):
            from_data = edge_data.get("from", {})
            to_data = edge_data.get("to", {})
            edge = Edge(
                from_task=TaskPort(task=from_data.get("task"), port=from_data.get("port")),
                to_task=TaskPort(task=to_data.get("task"), port=to_data.get("port")),
            )
            edges.append(edge)

        spec = PipelineSpec(
            api_version=data.get("apiVersion", "flowforge.io/v1"),
            kind=data.get("kind", "Pipeline"),
            metadata=metadata,
            tasks=tasks,
            edges=edges,
        )
        return spec
