"""A constrained, analyzable expression grammar.

Expressions come from intent/plan YAML (and ultimately the no-code UI), never as
arbitrary Python — so every expression's result type can be inferred statically
and lowered to a safe Ibis expression. This is what keeps ``derive`` and
``filter`` both trustworthy and portable across DuckDB and Spark.

Expression forms (JSON/YAML friendly):

    {"col": "amount"}                                  # column reference
    {"lit": 5}                                          # literal
    {"op": "*", "args": [ {"col": "amount"}, {"lit": 2} ]}
    {"fn": "date_trunc", "args": [ {"lit": "day"}, {"col": "created_at"} ], "type": "date"}
"""

from __future__ import annotations

from typing import Any, Dict

from .schema import Schema

_ARITH = {"+", "-", "*", "/"}
_CMP = {"=", "!=", "<", "<=", ">", ">="}
_BOOL = {"and", "or", "not"}

# fn name -> result dtype (None means "infer from first argument")
_FUNCTIONS = {
    "upper": "string", "lower": "string", "trim": "string", "concat": "string",
    "length": "int", "abs": None, "round": "double",
    "coalesce": None, "cast": None, "date_trunc": "date",
}


def _lit_dtype(value: Any) -> str:
    if isinstance(value, bool):
        return "bool"
    if isinstance(value, int):
        return "long"
    if isinstance(value, float):
        return "double"
    return "string"


class ExprError(ValueError):
    pass


def infer_type(expr: Dict[str, Any], schema: Schema) -> str:
    """Infer the logical dtype of *expr* against *schema*, without executing."""
    if "col" in expr:
        return schema.require(expr["col"]).dtype
    if "lit" in expr:
        return _lit_dtype(expr["lit"])
    if "op" in expr:
        op = expr["op"]
        if op in _CMP or op in _BOOL:
            return "bool"
        if op in _ARITH:
            arg_types = [infer_type(a, schema) for a in expr["args"]]
            if "double" in arg_types or "decimal" in arg_types or op == "/":
                return "double"
            return "long"
        raise ExprError(f"unknown operator {op!r}")
    if "fn" in expr:
        fn = expr["fn"]
        if fn not in _FUNCTIONS:
            raise ExprError(f"function {fn!r} is not allowed")
        declared = expr.get("type")
        if declared:
            return declared
        result = _FUNCTIONS[fn]
        if result is not None:
            return result
        # infer-from-first-arg functions (abs, coalesce, cast)
        if fn == "cast":
            raise ExprError("cast requires an explicit 'type'")
        return infer_type(expr["args"][0], schema)
    raise ExprError(f"invalid expression: {expr!r}")


def infer_nullable(expr: Dict[str, Any], schema: Schema) -> bool:
    """Infer whether *expr* can produce null, statically.

    Conservative but useful: a value derived only from non-nullable columns and
    literals is non-nullable, so a derived grouping key from a required column
    correctly satisfies a non-nullable target column.
    """
    if "col" in expr:
        return schema.require(expr["col"]).nullable
    if "lit" in expr:
        return expr["lit"] is None
    if "op" in expr:
        return any(infer_nullable(a, schema) for a in expr["args"])
    if "fn" in expr:
        args = expr.get("args", [])
        if expr["fn"] == "coalesce":
            return all(infer_nullable(a, schema) for a in args)
        return any(infer_nullable(a, schema) for a in args)
    return True


def lower(expr: Dict[str, Any], table) -> Any:
    """Lower *expr* to an Ibis value expression against an Ibis *table*."""
    if "col" in expr:
        return table[expr["col"]]
    if "lit" in expr:
        import ibis

        return ibis.literal(expr["lit"])
    if "op" in expr:
        op = expr["op"]
        args = [lower(a, table) for a in expr["args"]]
        if op == "+":
            return args[0] + args[1]
        if op == "-":
            return args[0] - args[1]
        if op == "*":
            return args[0] * args[1]
        if op == "/":
            return args[0] / args[1]
        if op == "=":
            return args[0] == args[1]
        if op == "!=":
            return args[0] != args[1]
        if op == "<":
            return args[0] < args[1]
        if op == "<=":
            return args[0] <= args[1]
        if op == ">":
            return args[0] > args[1]
        if op == ">=":
            return args[0] >= args[1]
        if op == "and":
            return args[0] & args[1]
        if op == "or":
            return args[0] | args[1]
        if op == "not":
            return ~args[0]
        raise ExprError(f"unknown operator {op!r}")
    if "fn" in expr:
        fn = expr["fn"]
        args = expr.get("args", [])
        if fn == "date_trunc":
            unit = args[0]["lit"]
            value = lower(args[1], table)
            if unit == "day":
                return value.date()  # truncates to a real date type
            return value.truncate(unit[0].upper())
        lowered = [lower(a, table) for a in args]
        if fn == "upper":
            return lowered[0].upper()
        if fn == "lower":
            return lowered[0].lower()
        if fn == "trim":
            return lowered[0].strip()
        if fn == "length":
            return lowered[0].length()
        if fn == "abs":
            return lowered[0].abs()
        if fn == "round":
            return lowered[0].round()
        if fn == "concat":
            out = lowered[0]
            for part in lowered[1:]:
                out = out + part
            return out
        if fn == "coalesce":
            import ibis

            return ibis.coalesce(*lowered)
        if fn == "cast":
            from .schema import _DTYPE_TO_IBIS

            return lowered[0].cast(_DTYPE_TO_IBIS[expr["type"]])
        raise ExprError(f"function {fn!r} is not allowed")
    raise ExprError(f"invalid expression: {expr!r}")
