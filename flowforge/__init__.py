"""FlowForge: a local-first workbench that compiles one Python-authored
workflow to multiple orchestrators and reports what each target preserves and
drops.

Public authoring API:

    from flowforge import task, pipeline
"""

from __future__ import annotations

from .authoring import Pipeline, TaskDef, pipeline, task
from .capability import analyze, analyze_all
from .compilers import compile as compile_target
from .ir import PipelineSpec
from .parser import parse_file, parse_module
from .validation import validate

__version__ = "0.2.0"

__all__ = [
    "task",
    "pipeline",
    "Pipeline",
    "TaskDef",
    "PipelineSpec",
    "parse_file",
    "parse_module",
    "validate",
    "analyze",
    "analyze_all",
    "compile_target",
    "__version__",
]
