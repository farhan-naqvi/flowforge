"""FlowForge Python SDK - Declarative data pipeline framework."""

__version__ = "0.1.0"
__author__ = "FlowForge Contributors"

from flowforge.core.task import task, Task, TaskConfig
from flowforge.core.pipeline import pipeline, Pipeline, PipelineConfig
from flowforge.decorators.common import (
    kafka, s3_read, sql_read,
    transform, aggregate, filter_data,
    save, s3_write, sql_write, notify,
)
from flowforge.executor.local import LocalExecutor
from flowforge.compiler.ir_exporter import pipeline_to_ir
from flowforge.graph.visualizer import DAGVisualizer
from flowforge.schema.inference import infer_schema, validate_data
from flowforge.cli.cli import CLI

# Re-export IR types for convenience
from flowforge.ir import (
    PipelineSpec,
    TaskType,
    Handler,
)

__all__ = [
    # Core
    'task', 'Task', 'TaskConfig',
    'pipeline', 'Pipeline', 'PipelineConfig',
    
    # Decorators
    'kafka', 's3_read', 'sql_read',
    'transform', 'aggregate', 'filter_data',
    'save', 's3_write', 'sql_write', 'notify',
    
    # Execution
    'LocalExecutor',
    
    # Compilation
    'pipeline_to_ir',
    
    # Visualization
    'DAGVisualizer',
    
    # Schema
    'infer_schema', 'validate_data',
    
    # CLI
    'CLI',
    
    # IR types
    'PipelineSpec', 'TaskType', 'Handler',
]
