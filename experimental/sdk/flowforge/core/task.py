"""Core task abstraction for FlowForge pipelines."""

from __future__ import annotations
from typing import Any, Callable, Dict, Optional, List, get_type_hints
from dataclasses import dataclass, field
import functools
import inspect
from uuid import uuid4


@dataclass
class TaskInputPort:
    """Input port specification."""
    name: str
    type_hint: Optional[type] = None
    description: str = ""


@dataclass
class TaskOutputPort:
    """Output port specification."""
    name: str
    type_hint: Optional[type] = None
    description: str = ""


@dataclass
class TaskConfig:
    """Task configuration."""
    image: Optional[str] = None
    command: Optional[List[str]] = None
    env: Dict[str, str] = field(default_factory=dict)
    resources: Dict[str, Any] = field(default_factory=dict)
    timeout: Optional[str] = None
    retries: int = 0


class Task:
    """Represents a task in a pipeline."""

    def __init__(
        self,
        func: Callable,
        task_id: Optional[str] = None,
        config: Optional[TaskConfig] = None,
        description: Optional[str] = None,
    ):
        """Initialize a task.
        
        Args:
            func: The function to wrap
            task_id: Unique task identifier (auto-generated if not provided)
            config: Task configuration
            description: Task description
        """
        self.func = func
        self.task_id = task_id or f"{func.__name__}_{uuid4().hex[:8]}"
        self.config = config or TaskConfig()
        self.description = description or func.__doc__ or ""
        self.name = func.__name__
        
        # Extract type hints for schema inference
        self.type_hints = get_type_hints(func) if hasattr(func, '__annotations__') else {}
        
        # Input/output ports
        self.inputs: Dict[str, TaskInputPort] = {}
        self.outputs: Dict[str, TaskOutputPort] = {}
        
        # Extract from function signature
        self._extract_ports()
        
        # Execution state (set at runtime)
        self._result: Optional[Any] = None
        self._executed = False

    def _extract_ports(self):
        """Extract input/output ports from function signature."""
        sig = inspect.signature(self.func)
        
        # Inputs: function parameters
        for param_name, param in sig.parameters.items():
            type_hint = self.type_hints.get(param_name)
            self.inputs[param_name] = TaskInputPort(
                name=param_name,
                type_hint=type_hint,
            )
        
        # Outputs: return type
        if 'return' in self.type_hints and self.type_hints['return'] is not None:
            return_type = self.type_hints['return']
            self.outputs['result'] = TaskOutputPort(
                name='result',
                type_hint=return_type,
            )

    def __call__(self, *args, **kwargs) -> Any:
        """Execute the task function.
        
        This is called during pipeline execution.
        """
        result = self.func(*args, **kwargs)
        self._result = result
        self._executed = True
        return result

    def get_result(self) -> Optional[Any]:
        """Get the last execution result."""
        return self._result

    def has_executed(self) -> bool:
        """Check if task has been executed."""
        return self._executed

    def reset(self):
        """Reset execution state."""
        self._result = None
        self._executed = False

    def __repr__(self) -> str:
        return f"<Task {self.name} id={self.task_id}>"


class TaskReference:
    """Reference to a task output during pipeline definition.
    
    This allows connecting task outputs to inputs in a Pythonic way.
    """

    def __init__(self, task: Task, port: str = "result"):
        """Initialize a task reference.
        
        Args:
            task: The source task
            port: The output port name
        """
        self.task = task
        self.port = port
        self._value: Optional[Any] = None

    def __call__(self, *args, **kwargs) -> Any:
        """Make the reference callable for execution context."""
        return self._value

    def __repr__(self) -> str:
        return f"<TaskRef {self.task.name}.{self.port}>"


def task(
    *,
    task_id: Optional[str] = None,
    image: Optional[str] = None,
    command: Optional[List[str]] = None,
    env: Optional[Dict[str, str]] = None,
    resources: Optional[Dict[str, Any]] = None,
    timeout: Optional[str] = None,
    retries: int = 0,
    description: Optional[str] = None,
) -> Callable:
    """Decorator to mark a function as a pipeline task.
    
    Example:
        @task(image="python:3.11")
        def extract(kafka_topic: str) -> list:
            '''Extract from Kafka.'''
            ...
    
    Args:
        task_id: Unique task identifier (auto-generated if not provided)
        image: Docker image for execution
        command: Command to run in container
        env: Environment variables
        resources: Resource requests/limits (cpu, memory)
        timeout: Task timeout
        retries: Number of retries on failure
        description: Task description
    
    Returns:
        Decorated function (Task object)
    """
    def decorator(func: Callable) -> Task:
        config = TaskConfig(
            image=image,
            command=command,
            env=env or {},
            resources=resources or {},
            timeout=timeout,
            retries=retries,
        )
        task_obj = Task(
            func=func,
            task_id=task_id,
            config=config,
            description=description,
        )
        
        # Preserve function metadata
        functools.update_wrapper(task_obj, func)
        
        return task_obj
    
    # Handle both @task and @task() syntax
    if len(task.__code__.co_varnames) > 0 and callable(task.__code__.co_consts[0]):
        # Called as @task(args)
        return decorator
    else:
        # Called as @task
        return decorator(task)
