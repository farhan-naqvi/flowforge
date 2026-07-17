"""ODCS (Open Data Contract Standard) loading.

We parse an ODCS-shaped YAML contract into two things Covenant needs:
- a ``Schema`` (typed, nullable fields) — the currency of plan-time inference;
- a set of ``quality`` rules (primary key uniqueness, required/not-null columns).

We support the subset of ODCS v3 that maps cleanly to those needs; unknown keys
are ignored rather than rejected, so real-world contracts load without fuss.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import List

import yaml

from covenant_transforms.schema import Field, Schema

# ODCS logicalType -> Covenant dtype
_LOGICAL = {
    "string": "string", "text": "string",
    "integer": "long", "int": "long", "bigint": "long",
    "number": "double", "float": "double", "double": "double",
    "decimal": "decimal",
    "boolean": "bool", "bool": "bool",
    "date": "date",
    "timestamp": "timestamp", "datetime": "timestamp",
}


@dataclass
class Contract:
    id: str
    name: str
    schema: Schema
    primary_key: List[str] = field(default_factory=list)
    required: List[str] = field(default_factory=list)

    @property
    def quality_rules(self) -> List[dict]:
        rules: List[dict] = []
        if self.primary_key:
            rules.append({"rule": "unique", "columns": self.primary_key})
        for col in self.required:
            rules.append({"rule": "not_null", "column": col})
        rules.append({"rule": "row_count_positive"})
        return rules


def _object_schema(obj: dict) -> tuple:
    props = obj.get("properties") or obj.get("fields") or []
    fields: List[Field] = []
    pk: List[str] = []
    required: List[str] = []
    for p in props:
        name = p["name"]
        logical = str(p.get("logicalType") or p.get("type") or "string").lower()
        dtype = _LOGICAL.get(logical, "string")
        is_required = bool(p.get("required", False))
        nullable = not is_required
        fields.append(Field(name, dtype, nullable))
        if is_required:
            required.append(name)
        if p.get("primaryKey") or p.get("unique"):
            pk.append(name)
    return Schema(tuple(fields)), pk, required


def load_contract(path: str) -> Contract:
    with open(path, "r", encoding="utf-8") as fh:
        doc = yaml.safe_load(fh)
    return parse_contract(doc)


def parse_contract(doc: dict) -> Contract:
    objects = doc.get("schema") or []
    if not objects:
        raise ValueError("ODCS contract has no 'schema' objects")
    # v1: one table object per contract (warehouse-shaped).
    obj = objects[0]
    schema, pk, required = _object_schema(obj)
    # A contract-level primaryKey list overrides per-property flags if present.
    if obj.get("primaryKey"):
        pk = list(obj["primaryKey"])
    return Contract(
        id=str(doc.get("id", doc.get("name", "contract"))),
        name=str(obj.get("name", doc.get("name", "table"))),
        schema=schema,
        primary_key=pk,
        required=required,
    )
