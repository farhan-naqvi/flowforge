"""Primitive base class and registry.

A primitive is a declarative, versioned transform: it knows how to (a) infer its
output schema from input schemas + params, and (b) lower to an Ibis expression.
It contains no backend-specific code — Ibis compiles the same lowering to DuckDB
(local verify) and Spark (full scale).
"""

from __future__ import annotations

from typing import Dict, List

from .schema import Schema


class Primitive:
    id: str = ""
    version: str = "0.1.0"
    arity: int = 1  # number of input tables (2 for join/union)

    def infer(self, inputs: List[Schema], params: dict) -> Schema:
        raise NotImplementedError

    def lower(self, inputs: list, params: dict):
        """Lower to an Ibis table expression given Ibis input tables."""
        raise NotImplementedError

    # small helper for single-input primitives
    def _one(self, inputs, kind="schema"):
        if len(inputs) != 1:
            raise ValueError(f"{self.id} expects 1 input, got {len(inputs)}")
        return inputs[0]


_REGISTRY: Dict[str, Primitive] = {}


def register(cls):
    """Class decorator: instantiate the primitive and register the instance.

    Returns the class so the decorated name still refers to the type.
    """
    _REGISTRY[cls.id] = cls()
    return cls


def get(primitive_id: str) -> Primitive:
    if primitive_id not in _REGISTRY:
        raise KeyError(f"unknown primitive {primitive_id!r}; known: {', '.join(sorted(_REGISTRY))}")
    return _REGISTRY[primitive_id]


def all_ids() -> List[str]:
    return sorted(_REGISTRY)
