"""Deterministic synthetic test-data generation from a source contract.

Produces rows that (a) conform to the source schema and (b) include deliberate
edge cases — nulls in nullable columns, duplicate primary keys, and boundary
values — so the local verify exercises the pipeline the way real data would.

This is the deterministic default; an LLM "test-data agent" can enrich these
cases later, but the generated fixture is always checked by the same verifier.
"""

from __future__ import annotations

import datetime as _dt
from typing import Dict, List

from covenant_transforms.schema import Schema

from .odcs import Contract


def _value(dtype: str, i: int):
    if dtype in ("int", "long"):
        return i
    if dtype in ("double", "decimal"):
        return float(i) + 0.5
    if dtype == "bool":
        return i % 2 == 0
    if dtype == "date":
        return _dt.date(2026, 1, 1) + _dt.timedelta(days=i % 5)
    if dtype == "timestamp":
        # spread across several days so daily rollups produce multiple rows
        return _dt.datetime(2026, 1, 1) + _dt.timedelta(hours=i * 8)
    return f"val_{i}"


def generate(contract: Contract, rows: int = 12) -> Dict[str, List]:
    """Return a column-oriented dict suitable for ``ibis.memtable``."""
    schema: Schema = contract.schema
    pk = set(contract.primary_key)
    data: Dict[str, List] = {f.name: [] for f in schema.fields}

    for i in range(rows):
        for f in schema.fields:
            # Edge case: a null in a nullable, non-PK column on ~every 5th row.
            if f.nullable and f.name not in pk and i % 5 == 4:
                data[f.name].append(None)
            # Edge case: duplicate a PK value once, to exercise dedup/uniqueness.
            elif f.name in pk and i == rows - 1:
                data[f.name].append(data[f.name][0])
            else:
                data[f.name].append(_value(f.dtype, i))

    # Bias a 'status'-like string column so filters have something to remove.
    for name in data:
        f = schema.get(name)
        if f and f.dtype == "string" and name.lower() in ("status", "state", "type"):
            data[name] = [["completed", "pending"][i % 2] for i in range(rows)]
    return data
