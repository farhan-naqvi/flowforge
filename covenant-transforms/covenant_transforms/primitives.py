"""The v1 primitive set for S3 Delta -> Delta (warehouse-shaped) transforms.

Each primitive implements ``infer`` (schema, static) and ``lower`` (Ibis). Delta
I/O is modelled by ``source``/``sink`` — the actual read/write is bound by the
runner (local: a DuckDB table; full-scale: a Delta path on Spark), so the
transform logic in between is identical across backends.
"""

from __future__ import annotations

from typing import List

from . import expr
from .primitive import Primitive, register
from .schema import DTYPES, Field, Schema

_AGG_RESULT = {"sum": None, "avg": "double", "min": None, "max": None, "count": "long"}


@register
class Source(Primitive):
    """Bind a named input table. Its schema is declared (from the source contract)."""

    id, arity = "source", 0

    def infer(self, inputs: List[Schema], params: dict) -> Schema:
        return Schema.from_dicts(params["schema"])

    def lower(self, inputs: list, params: dict):
        # The runner binds params["name"] -> an Ibis table before lowering.
        return params["_bound_table"]


@register
class Sink(Primitive):
    """Terminal write. Passes the schema through; conformance is checked against
    the target contract by the verifier."""

    id = "sink"

    def infer(self, inputs, params):
        return self._one(inputs)

    def lower(self, inputs, params):
        return inputs[0]


@register
class Select(Primitive):
    id = "select"

    def infer(self, inputs, params):
        return self._one(inputs).select(params["columns"])

    def lower(self, inputs, params):
        return inputs[0].select(params["columns"])


@register
class Drop(Primitive):
    id = "drop"

    def infer(self, inputs, params):
        return self._one(inputs).drop(params["columns"])

    def lower(self, inputs, params):
        return inputs[0].drop(*params["columns"])


@register
class Rename(Primitive):
    id = "rename"

    def infer(self, inputs, params):
        return self._one(inputs).rename(params["map"])

    def lower(self, inputs, params):
        # ibis rename takes new_name=old_name
        return inputs[0].rename({new: old for old, new in params["map"].items()})


@register
class Cast(Primitive):
    id = "cast"

    def infer(self, inputs, params):
        schema = self._one(inputs)
        for col, dtype in params["casts"].items():
            if dtype not in DTYPES:
                raise ValueError(f"cast: unknown dtype {dtype!r}")
            f = schema.require(col)
            schema = schema.with_field(Field(col, dtype, f.nullable))
        return schema

    def lower(self, inputs, params):
        from .schema import _DTYPE_TO_IBIS

        t = inputs[0]
        return t.mutate(**{c: t[c].cast(_DTYPE_TO_IBIS[d]) for c, d in params["casts"].items()})


@register
class Filter(Primitive):
    id = "filter"

    def infer(self, inputs, params):
        schema = self._one(inputs)
        if expr.infer_type(params["predicate"], schema) != "bool":
            raise ValueError("filter predicate must be boolean")
        return schema

    def lower(self, inputs, params):
        return inputs[0].filter(expr.lower(params["predicate"], inputs[0]))


@register
class Derive(Primitive):
    """with_column: add or replace a column from a constrained expression."""

    id = "derive"

    def infer(self, inputs, params):
        schema = self._one(inputs)
        dtype = expr.infer_type(params["expr"], schema)
        nullable = expr.infer_nullable(params["expr"], schema)
        return schema.with_field(Field(params["column"], dtype, nullable))

    def lower(self, inputs, params):
        t = inputs[0]
        return t.mutate(**{params["column"]: expr.lower(params["expr"], t)})


@register
class Dedup(Primitive):
    id = "dedup"

    def infer(self, inputs, params):
        return self._one(inputs)

    def lower(self, inputs, params):
        keys = params.get("keys")
        return inputs[0].distinct(on=keys) if keys else inputs[0].distinct()


@register
class Aggregate(Primitive):
    id = "aggregate"

    def infer(self, inputs, params):
        schema = self._one(inputs)
        fields = [schema.require(g) for g in params.get("group_by", [])]
        for agg in params["aggs"]:
            fn = agg["fn"]
            if fn not in _AGG_RESULT:
                raise ValueError(f"aggregate: unknown fn {fn!r}")
            if fn == "count":
                fields.append(Field(agg["name"], "long", nullable=False))
                continue
            col = schema.require(agg["col"])
            dtype = _AGG_RESULT[fn] or col.dtype
            # sum/min/max/avg over a non-nullable column in a non-empty group is
            # non-nullable; inherit the source column's nullability.
            fields.append(Field(agg["name"], dtype, nullable=col.nullable))
        return Schema(tuple(fields))

    def lower(self, inputs, params):
        t = inputs[0]
        aggs = {}
        for agg in params["aggs"]:
            fn = agg["fn"]
            if fn == "count":
                aggs[agg["name"]] = t.count() if not params.get("group_by") else t[agg["col"]].count()
            else:
                aggs[agg["name"]] = getattr(t[agg["col"]], fn)()
        group_by = params.get("group_by", [])
        return t.group_by(group_by).aggregate(**aggs) if group_by else t.aggregate(**aggs)


@register
class Union(Primitive):
    id, arity = "union", 2

    def infer(self, inputs, params):
        if len(inputs) != 2:
            raise ValueError("union expects 2 inputs")
        if inputs[0].names != inputs[1].names:
            raise ValueError("union inputs must have the same columns")
        return inputs[0]

    def lower(self, inputs, params):
        return inputs[0].union(inputs[1])


@register
class Join(Primitive):
    id, arity = "join", 2

    def infer(self, inputs, params):
        left, right = inputs
        keys = params["on"]
        how = params.get("how", "inner")
        fields = list(left.fields)
        left_names = set(left.names)
        for f in right.fields:
            if f.name in keys:
                continue
            name = f.name if f.name not in left_names else f"{f.name}_right"
            nullable = f.nullable or how in ("left", "outer")
            fields.append(Field(name, f.dtype, nullable))
        return Schema(tuple(fields))

    def lower(self, inputs, params):
        left, right = inputs
        how = params.get("how", "inner")
        joined = left.join(right, predicates=params["on"], how=how)
        return joined
