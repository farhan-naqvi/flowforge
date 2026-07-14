"""FlowForge IR validators for Python."""

from abc import ABC, abstractmethod
from typing import List, Optional
from .spec import PipelineSpec, TaskPort


class Validator(ABC):
    """Base validator interface."""

    @abstractmethod
    def validate(self, spec: PipelineSpec) -> Optional[str]:
        """Validate the pipeline. Returns error message if invalid."""
        pass

    @abstractmethod
    def name(self) -> str:
        """Return validator name."""
        pass


class DAGValidator(Validator):
    """Validates that the pipeline DAG has no cycles."""

    def validate(self, spec: PipelineSpec) -> Optional[str]:
        """Check for cycles using DFS."""
        # Build adjacency list
        graph = {task_id: [] for task_id in spec.tasks}

        for edge in spec.edges:
            graph[edge.from_task.task].append(edge.to_task.task)

        # DFS cycle detection
        visited = set()
        rec_stack = set()

        def has_cycle_dfs(node: str, path: List[str]) -> bool:
            visited.add(node)
            rec_stack.add(node)
            path.append(node)

            for neighbor in graph.get(node, []):
                if neighbor not in visited:
                    if has_cycle_dfs(neighbor, path):
                        return True
                elif neighbor in rec_stack:
                    return True

            rec_stack.discard(node)
            path.pop()
            return False

        for task_id in spec.tasks:
            if task_id not in visited:
                path: List[str] = []
                if has_cycle_dfs(task_id, path):
                    return f"Cycle detected: {' -> '.join(path)}"

        return None

    def name(self) -> str:
        """Return validator name."""
        return "DAGValidator"


class SchemaValidator(Validator):
    """Validates that task input/output schemas are properly connected."""

    def validate(self, spec: PipelineSpec) -> Optional[str]:
        """Check schema validity."""
        for i, edge in enumerate(spec.edges):
            # Check tasks exist
            if edge.from_task.task not in spec.tasks:
                return (
                    f"Edge {i}: source task '{edge.from_task.task}' not found"
                )
            if edge.to_task.task not in spec.tasks:
                return f"Edge {i}: target task '{edge.to_task.task}' not found"

            # Check ports exist
            from_task = spec.tasks[edge.from_task.task]
            to_task = spec.tasks[edge.to_task.task]

            if edge.from_task.port not in from_task.outputs:
                return (
                    f"Edge {i}: output port '{edge.from_task.port}' "
                    f"not found on task '{edge.from_task.task}'"
                )
            if edge.to_task.port not in to_task.inputs:
                return (
                    f"Edge {i}: input port '{edge.to_task.port}' "
                    f"not found on task '{edge.to_task.task}'"
                )

        return None

    def name(self) -> str:
        """Return validator name."""
        return "SchemaValidator"


class CompositeValidator(Validator):
    """Combines multiple validators."""

    def __init__(self, validators: List[Validator]):
        """Initialize with list of validators."""
        self.validators = validators

    def validate(self, spec: PipelineSpec) -> Optional[str]:
        """Run all validators."""
        for validator in self.validators:
            error = validator.validate(spec)
            if error:
                return f"[{validator.name()}] {error}"
        return None

    def name(self) -> str:
        """Return validator name."""
        return "CompositeValidator"
