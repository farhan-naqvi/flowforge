"""FlowForge IR builder - fluent API for constructing pipelines."""

from typing import Dict, Any, Optional
from .spec import (
    PipelineSpec,
    PipelineMetadata,
    Task,
    TaskType,
    Handler,
    Edge,
    TaskPort,
    RetryPolicy,
    CostEstimate,
)


class PipelineBuilder:
    """Fluent builder for constructing pipelines programmatically."""

    def __init__(self, name: str):
        """Initialize a new pipeline builder."""
        self.spec = PipelineSpec(
            metadata=PipelineMetadata(name=name),
            tasks={},
            edges=[],
        )
        self._errors: list[str] = []

    def set_version(self, version: str) -> "PipelineBuilder":
        """Set the pipeline version."""
        if self.spec.metadata:
            self.spec.metadata.version = version
        return self

    def set_owner(self, owner: str) -> "PipelineBuilder":
        """Set the pipeline owner."""
        if self.spec.metadata:
            self.spec.metadata.owner = owner
        return self

    def set_description(self, description: str) -> "PipelineBuilder":
        """Set the pipeline description."""
        if self.spec.metadata:
            self.spec.metadata.description = description
        return self

    def add_tag(self, key: str, value: str) -> "PipelineBuilder":
        """Add a tag to the pipeline."""
        if self.spec.metadata:
            self.spec.metadata.tags[key] = value
        return self

    def add_task(
        self,
        task_id: str,
        task_type: TaskType,
        handler: Handler,
        description: Optional[str] = None,
    ) -> "PipelineBuilder":
        """Add a task to the pipeline."""
        if not task_id:
            self._errors.append("task ID cannot be empty")
            return self

        if task_id in self.spec.tasks:
            self._errors.append(f"task '{task_id}' already exists")
            return self

        task = Task(
            type=task_type,
            handler=handler,
            description=description,
        )
        self.spec.tasks[task_id] = task
        return self

    def add_input(
        self, task_id: str, port_name: str, schema: Dict[str, Any]
    ) -> "PipelineBuilder":
        """Add an input port to a task."""
        if task_id not in self.spec.tasks:
            self._errors.append(f"task '{task_id}' does not exist")
            return self

        self.spec.tasks[task_id].inputs[port_name] = schema
        return self

    def add_output(
        self, task_id: str, port_name: str, schema: Dict[str, Any]
    ) -> "PipelineBuilder":
        """Add an output port to a task."""
        if task_id not in self.spec.tasks:
            self._errors.append(f"task '{task_id}' does not exist")
            return self

        self.spec.tasks[task_id].outputs[port_name] = schema
        return self

    def add_edge(
        self, from_task: str, from_port: str, to_task: str, to_port: str
    ) -> "PipelineBuilder":
        """Connect an output port to an input port."""
        edge = Edge(
            from_task=TaskPort(task=from_task, port=from_port),
            to_task=TaskPort(task=to_task, port=to_port),
        )
        self.spec.edges.append(edge)
        return self

    def set_executor_config(
        self, task_id: str, executor: str, config: Dict[str, Any]
    ) -> "PipelineBuilder":
        """Set executor-specific configuration for a task."""
        if task_id not in self.spec.tasks:
            self._errors.append(f"task '{task_id}' does not exist")
            return self

        self.spec.tasks[task_id].executor_config[executor] = config
        return self

    def set_retry_policy(self, task_id: str, policy: RetryPolicy) -> "PipelineBuilder":
        """Set the retry policy for a task."""
        if task_id not in self.spec.tasks:
            self._errors.append(f"task '{task_id}' does not exist")
            return self

        self.spec.tasks[task_id].retry = policy
        return self

    def set_timeout(self, task_id: str, timeout: str) -> "PipelineBuilder":
        """Set the timeout for a task."""
        if task_id not in self.spec.tasks:
            self._errors.append(f"task '{task_id}' does not exist")
            return self

        self.spec.tasks[task_id].timeout = timeout
        return self

    def set_cost_estimate(
        self, task_id: str, estimate: CostEstimate
    ) -> "PipelineBuilder":
        """Set the cost estimate for a task."""
        if task_id not in self.spec.tasks:
            self._errors.append(f"task '{task_id}' does not exist")
            return self

        self.spec.tasks[task_id].cost_estimate = estimate
        return self

    def build(self) -> PipelineSpec:
        """Build and validate the pipeline."""
        if self._errors:
            raise ValueError(f"Builder errors: {self._errors}")

        self.spec.validate()
        return self.spec


# Convenience function
def pipeline(name: str) -> PipelineBuilder:
    """Create a new pipeline builder."""
    return PipelineBuilder(name)
