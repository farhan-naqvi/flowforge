"""Local pipeline executor for development and testing."""

from __future__ import annotations
from typing import Any, Dict, List, Optional, Set
from flowforge.core.pipeline import Pipeline
from flowforge.core.task import Task, TaskReference
import json
import os


class LocalExecutor:
    """Execute pipelines locally without external orchestrators."""

    def __init__(self, workers: int = 1, cache_dir: Optional[str] = None):
        """Initialize local executor.
        
        Args:
            workers: Number of parallel workers (for future use)
            cache_dir: Directory for caching task outputs
        """
        self.workers = workers
        self.cache_dir = cache_dir or "./.flowforge_cache"
        self.task_results: Dict[str, Any] = {}
        self.task_order: List[Task] = []

    def execute(self, pipeline_obj: Pipeline) -> Dict[str, Any]:
        """Execute pipeline locally.
        
        Args:
            pipeline_obj: Pipeline to execute
        
        Returns:
            Dictionary of task results {task_id: result}
        
        Raises:
            ValueError: If pipeline is invalid
            RuntimeError: If execution fails
        """
        # Validate pipeline
        errors = pipeline_obj.validate()
        if errors:
            raise ValueError(f"Pipeline validation failed: {errors}")
        
        # Create cache directory
        if not os.path.exists(self.cache_dir):
            os.makedirs(self.cache_dir)
        
        # Get execution order (topological sort)
        self.task_order = self._topological_sort(pipeline_obj)
        
        # Execute tasks
        self.task_results = {}
        for task in self.task_order:
            try:
                result = self._execute_task(pipeline_obj, task)
                self.task_results[task.task_id] = result
            except Exception as e:
                raise RuntimeError(f"Task '{task.name}' failed: {e}")
        
        return self.task_results

    def _execute_task(self, pipeline_obj: Pipeline, task: Task) -> Any:
        """Execute a single task.
        
        Args:
            pipeline_obj: The pipeline
            task: Task to execute
        
        Returns:
            Task result
        """
        # Resolve inputs from previous task outputs
        task_inputs = {}
        for from_task, from_port, to_task, to_port in pipeline_obj.edges:
            if to_task == task:
                if from_task.task_id in self.task_results:
                    task_inputs[to_port] = self.task_results[from_task.task_id]
        
        # Execute task function with resolved inputs
        if task_inputs:
            result = task(**task_inputs)
        else:
            # Source task with no inputs
            result = task()
        
        return result

    def _topological_sort(self, pipeline_obj: Pipeline) -> List[Task]:
        """Get tasks in topological order.
        
        Args:
            pipeline_obj: The pipeline
        
        Returns:
            List of tasks in execution order
        """
        in_degree: Dict[Task, int] = {task: 0 for task in pipeline_obj.tasks.values()}
        adj: Dict[Task, List[Task]] = {task: [] for task in pipeline_obj.tasks.values()}
        
        for from_task, _, to_task, _ in pipeline_obj.edges:
            adj[from_task].append(to_task)
            in_degree[to_task] += 1
        
        queue = [task for task in pipeline_obj.tasks.values() if in_degree[task] == 0]
        result = []
        
        while queue:
            queue.sort(key=lambda t: t.name)
            task = queue.pop(0)
            result.append(task)
            
            for neighbor in adj[task]:
                in_degree[neighbor] -= 1
                if in_degree[neighbor] == 0:
                    queue.append(neighbor)
        
        return result

    def get_result(self, task_id: str) -> Any:
        """Get result of a task.
        
        Args:
            task_id: Task identifier
        
        Returns:
            Task result
        """
        return self.task_results.get(task_id)


__all__ = ['LocalExecutor']
