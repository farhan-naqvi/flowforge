"""Local verify — the run-time half of the trust story.

Executes a plan on DuckDB (via Ibis) over synthetic test data and checks the
*actual* output against the target contract: schema conformance plus the
contract's quality rules (primary-key uniqueness, required-column not-null,
non-empty output). Because primitives lower through Ibis, this DuckDB run
predicts the full-scale Spark run — the same lowering, a smaller engine.
"""

from __future__ import annotations

import threading
from dataclasses import dataclass, field
from typing import Dict, List, Optional

import ibis

# DuckDB/Ibis execution is serialized: a single embedded DuckDB is not safe
# under concurrent access from multiple request threads. Local verify is fast,
# so a process-wide lock is the simplest correct choice.
_EXEC_LOCK = threading.Lock()

from covenant_transforms import get as get_primitive
from covenant_transforms.schema import Schema

from .model import Plan
from .odcs import Contract, load_contract
from .testdata import generate


@dataclass
class Verdict:
    ok: bool
    schema_problems: List[str] = field(default_factory=list)
    quality_problems: List[str] = field(default_factory=list)
    row_count: int = 0
    sample: List[dict] = field(default_factory=list)

    def to_dict(self) -> dict:
        return {
            "ok": self.ok,
            "schema_problems": self.schema_problems,
            "quality_problems": self.quality_problems,
            "row_count": self.row_count,
            "sample": self.sample,
        }


def _execute(plan: Plan, source_table) -> "ibis.Table":
    """Lower and chain every step; return the final Ibis table expression."""
    current = None
    for step in plan.steps:
        prim = get_primitive(step.primitive)
        if step.primitive == "source":
            params = dict(step.params)
            params["_bound_table"] = source_table
            current = prim.lower([], params)
        else:
            current = prim.lower([current], step.params)
    return current


def _produced_schema(ibis_schema, df) -> Schema:
    from covenant_transforms.schema import Field

    typed = Schema.from_ibis(ibis_schema)
    fields = []
    for f in typed.fields:
        actual_nullable = bool(df[f.name].isnull().any()) if f.name in df.columns else f.nullable
        fields.append(Field(f.name, f.dtype, actual_nullable))
    return Schema(tuple(fields))


def _check_quality(df, contract: Contract) -> List[str]:
    problems: List[str] = []
    for rule in contract.quality_rules:
        if rule["rule"] == "row_count_positive" and len(df) == 0:
            problems.append("output is empty (row_count_positive)")
        elif rule["rule"] == "unique":
            cols = [c for c in rule["columns"] if c in df.columns]
            if cols and df.duplicated(subset=cols).any():
                problems.append(f"primary key not unique: {', '.join(cols)}")
        elif rule["rule"] == "not_null":
            col = rule["column"]
            if col in df.columns and df[col].isnull().any():
                problems.append(f"required column '{col}' contains nulls")
    return problems


def verify_plan(
    plan: Plan,
    target: Contract,
    test_data: Optional[Dict[str, list]] = None,
    source: Optional[Contract] = None,
    base_dir: str = ".",
) -> Verdict:
    if source is None:
        import os

        source = load_contract(os.path.join(base_dir, plan.source_contract))
    if test_data is None:
        test_data = generate(source)

    with _EXEC_LOCK:
        con = None
        try:
            source_table = ibis.memtable(test_data)
            result = _execute(plan, source_table)
            con = ibis.duckdb.connect()
            df = con.to_pandas(result)
        except Exception as exc:  # noqa: BLE001 - execution failure is a real verdict
            return Verdict(ok=False, quality_problems=[f"execution failed: {exc}"])
        finally:
            if con is not None:
                con.disconnect()

    # Types come from the engine; nullability comes from the DATA (engines report
    # every column as nullable, so a schema-level nullability check would be a
    # false positive — the not_null quality rules verify it against real values).
    produced = _produced_schema(result.schema(), df)
    schema_problems = produced.conformance(target.schema)
    quality_problems = _check_quality(df, target)

    sample = df.head(5).to_dict(orient="records")
    ok = not schema_problems and not quality_problems
    return Verdict(
        ok=ok,
        schema_problems=schema_problems,
        quality_problems=quality_problems,
        row_count=int(len(df)),
        sample=_jsonable(sample),
    )


def _jsonable(rows: List[dict]) -> List[dict]:
    out = []
    for r in rows:
        out.append({k: (str(v) if not isinstance(v, (int, float, str, bool, type(None))) else v)
                    for k, v in r.items()})
    return out
