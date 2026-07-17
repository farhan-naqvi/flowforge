"""Executor-specific compilers. Each consumes canonical IR and emits a
deployable artifact using a safe serializer (never raw string interpolation
of user-controlled values)."""

from __future__ import annotations

from typing import Callable, Dict

from ..ir import PipelineSpec
from . import airflow, argo

_COMPILERS: Dict[str, Callable[[PipelineSpec], str]] = {
    "argo": argo.compile,
    "airflow": airflow.compile,
}

TARGETS = tuple(_COMPILERS)


def compile(spec: PipelineSpec, target: str) -> str:
    if target not in _COMPILERS:
        raise ValueError(f"unknown target {target!r}; known: {', '.join(TARGETS)}")
    return _COMPILERS[target](spec)
