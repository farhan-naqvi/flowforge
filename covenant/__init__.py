"""Covenant — a contract-first, verify-before-run pipeline platform.

Define ODCS source + target contracts, describe a pipeline as structured intent,
let the planner compose it from the ``covenant-transforms`` primitive library,
and verify it against the target contract — statically (schema inference) and on
synthetic data locally (DuckDB) — before compiling it to an Argo/Spark run.

The moat is trust (verify-before-run), not the code generation.
"""

from __future__ import annotations

from .compile import compile_argo
from .model import Intent, Plan, Step
from .odcs import Contract, load_contract
from .planner import DeterministicPlanner, plan_from_intent
from .verify import Verdict, verify_plan

__version__ = "0.1.0"

__all__ = [
    "Contract",
    "load_contract",
    "Intent",
    "Plan",
    "Step",
    "DeterministicPlanner",
    "plan_from_intent",
    "Verdict",
    "verify_plan",
    "compile_argo",
    "__version__",
]
