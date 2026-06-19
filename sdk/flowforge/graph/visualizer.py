"""DAG visualization for pipelines."""

from __future__ import annotations
from typing import List, Dict, Set
from flowforge.core.pipeline import Pipeline
from flowforge.core.task import Task


class DAGVisualizer:
    """Visualize pipeline DAG as ASCII art."""

    @staticmethod
    def visualize(pipeline_obj: Pipeline) -> str:
        """Generate ASCII representation of pipeline DAG.
        
        Args:
            pipeline_obj: The pipeline to visualize
        
        Returns:
            ASCII string representation
        """
        if not pipeline_obj.tasks:
            return "Empty pipeline"
        
        # Build adjacency info
        lines = []
        lines.append(f"Pipeline: {pipeline_obj.name}")
        lines.append(f"Tasks: {len(pipeline_obj.tasks)}, Edges: {len(pipeline_obj.edges)}")
        lines.append("")
        
        # Topological sort for better visualization
        sorted_tasks = DAGVisualizer._topological_sort(pipeline_obj)
        
        # Display tasks
        lines.append("Tasks:")
        for i, task in enumerate(sorted_tasks, 1):
            inputs = list(task.inputs.keys()) if task.inputs else ["(no inputs)"]
            outputs = list(task.outputs.keys()) if task.outputs else ["(no outputs)"]
            lines.append(f"  {i}. {task.name}")
            lines.append(f"     Inputs: {', '.join(inputs)}")
            lines.append(f"     Outputs: {', '.join(outputs)}")
        
        lines.append("")
        lines.append("Edges:")
        for from_task, from_port, to_task, to_port in pipeline_obj.edges:
            lines.append(f"  {from_task.name}.{from_port} → {to_task.name}.{to_port}")
        
        return "\n".join(lines)

    @staticmethod
    def to_graphviz(pipeline_obj: Pipeline) -> str:
        """Generate Graphviz DOT notation for the DAG.
        
        Args:
            pipeline_obj: The pipeline to visualize
        
        Returns:
            Graphviz DOT string
        """
        lines = ["digraph {"]
        lines.append(f'  label = "{pipeline_obj.name}";')
        lines.append("  rankdir = LR;")
        lines.append("")
        
        # Add nodes
        for task in pipeline_obj.tasks.values():
            lines.append(f'  "{task.task_id}" [label="{task.name}"];')
        
        lines.append("")
        
        # Add edges
        for from_task, _, to_task, _ in pipeline_obj.edges:
            lines.append(f'  "{from_task.task_id}" -> "{to_task.task_id}";')
        
        lines.append("}")
        return "\n".join(lines)

    @staticmethod
    def _topological_sort(pipeline_obj: Pipeline) -> List[Task]:
        """Get tasks in topological order.
        
        Args:
            pipeline_obj: The pipeline
        
        Returns:
            List of tasks in topological order
        """
        # Build adjacency and in-degree
        in_degree: Dict[Task, int] = {task: 0 for task in pipeline_obj.tasks.values()}
        adj: Dict[Task, List[Task]] = {task: [] for task in pipeline_obj.tasks.values()}
        
        for from_task, _, to_task, _ in pipeline_obj.edges:
            adj[from_task].append(to_task)
            in_degree[to_task] += 1
        
        # Kahn's algorithm
        queue = [task for task in pipeline_obj.tasks.values() if in_degree[task] == 0]
        result = []
        
        while queue:
            # Sort for consistency
            queue.sort(key=lambda t: t.name)
            task = queue.pop(0)
            result.append(task)
            
            for neighbor in adj[task]:
                in_degree[neighbor] -= 1
                if in_degree[neighbor] == 0:
                    queue.append(neighbor)
        
        return result


__all__ = ['DAGVisualizer']
