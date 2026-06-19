"""Export Python Pipeline to FlowForge IR (PipelineSpec)."""

from __future__ import annotations
from typing import Dict, Any, Optional
from flowforge.ir import (
    PipelineSpec,
    PipelineMetadata,
    Task as IRTask,
    TaskType,
    Handler,
    Edge,
    TaskPort,
    RetryPolicy,
    CostEstimate,
)
from flowforge.core.pipeline import Pipeline
from flowforge.core.task import Task
from flowforge.schema.inference import infer_schema


class IRExporter:
    """Export Python Pipeline to FlowForge IR."""

    @staticmethod
    def export(pipeline_obj: Pipeline) -> PipelineSpec:
        """Export a Pipeline to PipelineSpec.
        
        Args:
            pipeline_obj: The Pipeline to export
        
        Returns:
            PipelineSpec (IR representation)
        """
        # Create metadata
        metadata = PipelineMetadata(
            name=pipeline_obj.config.name,
            version=pipeline_obj.config.version,
            owner=pipeline_obj.config.owner,
            description=pipeline_obj.description,
            tags=pipeline_obj.config.tags,
        )
        
        # Convert tasks
        ir_tasks: Dict[str, IRTask] = {}
        for task_id, task in pipeline_obj.tasks.items():
            ir_task = IRExporter._convert_task(task_id, task)
            ir_tasks[task_id] = ir_task
        
        # Convert edges
        ir_edges = []
        for from_task, from_port, to_task, to_port in pipeline_obj.edges:
            edge = Edge(
                from_task=TaskPort(task=from_task.task_id, port=from_port),
                to_task=TaskPort(task=to_task.task_id, port=to_port),
            )
            ir_edges.append(edge)
        
        # Create PipelineSpec
        spec = PipelineSpec(
            metadata=metadata,
            tasks=ir_tasks,
            edges=ir_edges,
        )
        
        return spec

    @staticmethod
    def _convert_task(task_id: str, task: Task) -> IRTask:
        """Convert a Task to IR Task.
        
        Args:
            task_id: Task identifier
            task: The Python Task
        
        Returns:
            IR Task
        """
        # Infer handler type (default to Python)
        handler = Handler(
            type=task.config.command[0] if task.config.command else "python",
            source=task.name,  # Function name as source
        )
        
        # Infer input/output schemas
        inputs = {}
        for port_name, port in task.inputs.items():
            inputs[port_name] = infer_schema(port.type_hint)
        
        outputs = {}
        for port_name, port in task.outputs.items():
            outputs[port_name] = infer_schema(port.type_hint)
        
        # Build executor config
        executor_config = {}
        if task.config.image:
            executor_config["argo"] = {"image": task.config.image}
            executor_config["airflow"] = {}
        
        # Build retry policy
        retry_policy = None
        if task.config.retries > 0:
            retry_policy = RetryPolicy(
                max_attempts=task.config.retries + 1,
                backoff="exponential",
                backoff_multiplier=2.0,
                initial_delay_seconds=1,
            )
        
        # Create IR Task
        ir_task = IRTask(
            type=TaskType.TRANSFORM,  # Default type (could be inferred from task name)
            handler=handler,
            description=task.description,
            inputs=inputs,
            outputs=outputs,
            executor_config=executor_config,
            retry=retry_policy,
            timeout=task.config.timeout,
        )
        
        return ir_task


def pipeline_to_ir(pipeline_obj: Pipeline) -> PipelineSpec:
    """Convenience function to export pipeline to IR.
    
    Args:
        pipeline_obj: Python Pipeline object
    
    Returns:
        PipelineSpec
    """
    return IRExporter.export(pipeline_obj)


__all__ = ['IRExporter', 'pipeline_to_ir']
