"""Pipeline abstraction and @pipeline decorator."""

from __future__ import annotations
from typing import Any, Callable, Dict, List, Optional, Set
from dataclasses import dataclass, field
import functools
import inspect
from .task import Task, TaskReference


@dataclass
class PipelineConfig:
    """Pipeline configuration."""
    name: str
    version: str = "0.1.0"
    owner: Optional[str] = None
    description: Optional[str] = None
    tags: Dict[str, str] = field(default_factory=dict)


class Pipeline:
    """Represents a FlowForge pipeline."""

    def __init__(
        self,
        func: Callable,
        config: Optional[PipelineConfig] = None,
    ):
        """Initialize a pipeline.
        
        Args:
            func: The function defining the pipeline
            config: Pipeline configuration
        """
        self.func = func
        self.name = func.__name__
        self.config = config or PipelineConfig(name=func.__name__)
        self.description = func.__doc__ or ""
        
        # Tasks defined in this pipeline
        self.tasks: Dict[str, Task] = {}
        
        # Edges (connections between tasks)
        self.edges: List[tuple[Task, str, Task, str]] = []
        
        # Local variables during pipeline definition
        self._local_vars: Dict[str, Any] = {}
        
        # Build the pipeline when created
        self._build()

    def _build(self):
        """Build the pipeline by executing the decorated function.
        
        This analyzes the function to detect task definitions and edges.
        """
        # Create a custom frame to track variable assignments
        frame_locals: Dict[str, Any] = {}
        
        # Execute the function with our frame
        try:
            self.func.__globals__.update({'_pipeline_context': self})
            exec(
                inspect.getsource(self.func),
                {**self.func.__globals__, '_pipeline_context': self},
                frame_locals,
            )
        except Exception:
            # If exec fails, try simple execution
            pass
        
        self._local_vars = frame_locals

    def add_task(self, task_id: str, task: Task) -> None:
        """Register a task in the pipeline.
        
        Args:
            task_id: Unique task identifier
            task: The Task object
        """
        if task_id in self.tasks:
            raise ValueError(f"Task '{task_id}' already exists")
        self.tasks[task_id] = task

    def add_edge(
        self,
        from_task: Task,
        from_port: str,
        to_task: Task,
        to_port: str,
    ) -> None:
        """Connect two tasks.
        
        Args:
            from_task: Source task
            from_port: Output port name
            to_task: Target task
            to_port: Input port name
        """
        if from_port not in from_task.outputs:
            raise ValueError(
                f"Output port '{from_port}' not found on task {from_task.name}"
            )
        if to_port not in to_task.inputs:
            raise ValueError(
                f"Input port '{to_port}' not found on task {to_task.name}"
            )
        
        self.edges.append((from_task, from_port, to_task, to_port))

    def get_tasks(self) -> Dict[str, Task]:
        """Get all tasks in the pipeline."""
        return self.tasks.copy()

    def get_edges(self) -> List[tuple[Task, str, Task, str]]:
        """Get all edges in the pipeline."""
        return self.edges.copy()

    def get_source_tasks(self) -> List[Task]:
        """Get tasks with no predecessors (sources)."""
        predecessors = set()
        for _, _, to_task, _ in self.edges:
            predecessors.add(to_task)
        
        sources = []
        for task in self.tasks.values():
            if task not in predecessors:
                sources.append(task)
        return sources

    def get_sink_tasks(self) -> List[Task]:
        """Get tasks with no successors (sinks)."""
        successors = set()
        for from_task, _, _, _ in self.edges:
            successors.add(from_task)
        
        sinks = []
        for task in self.tasks.values():
            if task not in successors:
                sinks.append(task)
        return sinks

    def validate(self) -> List[str]:
        """Validate the pipeline.
        
        Returns:
            List of validation errors (empty if valid)
        """
        errors = []
        
        # Check for empty pipeline
        if not self.tasks:
            errors.append("Pipeline has no tasks")
        
        # Check for unreachable tasks
        reachable = self._get_reachable_tasks()
        for task in self.tasks.values():
            if task not in reachable:
                errors.append(f"Task '{task.name}' is unreachable")
        
        # Check edges reference valid tasks
        for from_task, from_port, to_task, to_port in self.edges:
            if from_port not in from_task.outputs:
                errors.append(
                    f"Edge: output port '{from_port}' not on task '{from_task.name}'"
                )
            if to_port not in to_task.inputs:
                errors.append(
                    f"Edge: input port '{to_port}' not on task '{to_task.name}'"
                )
        
        return errors

    def _get_reachable_tasks(self) -> Set[Task]:
        """Get all reachable tasks from source tasks."""
        reachable = set()
        queue = self.get_source_tasks()
        
        while queue:
            task = queue.pop(0)
            if task in reachable:
                continue
            reachable.add(task)
            
            # Add successors
            for from_task, _, to_task, _ in self.edges:
                if from_task == task and to_task not in reachable:
                    queue.append(to_task)
        
        return reachable

    def __repr__(self) -> str:
        return f"<Pipeline {self.name} with {len(self.tasks)} tasks>"


def pipeline(
    *,
    name: Optional[str] = None,
    version: str = "0.1.0",
    owner: Optional[str] = None,
    description: Optional[str] = None,
    tags: Optional[Dict[str, str]] = None,
) -> Callable:
    """Decorator to define a pipeline.
    
    Example:
        @pipeline(name="etl", version="1.0.0", owner="data_team")
        def my_pipeline():
            source = kafka("events")
            clean = transform(source, image="clean:v1")
            save(clean)
    
    Args:
        name: Pipeline name (defaults to function name)
        version: Semantic version
        owner: Pipeline owner/team
        description: Pipeline description
        tags: Key-value tags for organizing pipelines
    
    Returns:
        Pipeline object
    """
    def decorator(func: Callable) -> Pipeline:
        config = PipelineConfig(
            name=name or func.__name__,
            version=version,
            owner=owner,
            description=description,
            tags=tags or {},
        )
        pipeline_obj = Pipeline(func, config)
        
        # Preserve function metadata
        functools.update_wrapper(pipeline_obj, func)
        
        return pipeline_obj
    
    return decorator
