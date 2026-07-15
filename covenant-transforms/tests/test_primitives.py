import os
import sys

import ibis

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from covenant_transforms import get  # noqa: E402
from covenant_transforms.schema import Field, Schema  # noqa: E402

ORDERS = Schema.of(
    Field("order_id", "long", False),
    Field("customer_id", "long", True),
    Field("amount", "double", True),
    Field("status", "string", True),
)


def _table():
    return ibis.memtable(
        {
            "order_id": [1, 2, 3, 4],
            "customer_id": [10, 10, 20, 20],
            "amount": [5.0, 15.0, 25.0, 35.0],
            "status": ["completed", "pending", "completed", "completed"],
        }
    )


def _run(t):
    return ibis.duckdb.connect().to_pandas(t)


def test_filter_infers_and_runs():
    p = get("filter")
    params = {"predicate": {"op": "=", "args": [{"col": "status"}, {"lit": "completed"}]}}
    out_schema = p.infer([ORDERS], params)
    assert out_schema.names == ORDERS.names  # filter preserves schema
    df = _run(p.lower([_table()], params))
    assert len(df) == 3 and set(df["status"]) == {"completed"}


def test_derive_infers_new_column_type():
    p = get("derive")
    params = {"column": "double_amount", "expr": {"op": "*", "args": [{"col": "amount"}, {"lit": 2}]}}
    out = p.infer([ORDERS], params)
    assert out.require("double_amount").dtype == "double"
    df = _run(p.lower([_table()], params))
    assert df.loc[df["order_id"] == 1, "double_amount"].iloc[0] == 10.0


def test_aggregate_schema_and_values():
    p = get("aggregate")
    params = {
        "group_by": ["customer_id"],
        "aggs": [
            {"name": "total", "fn": "sum", "col": "amount"},
            {"name": "n", "fn": "count", "col": "order_id"},
        ],
    }
    out = p.infer([ORDERS], params)
    assert out.names == ["customer_id", "total", "n"]
    assert out.require("total").dtype == "double"
    assert out.require("n").dtype == "long"
    df = _run(p.lower([_table()], params)).sort_values("customer_id").reset_index(drop=True)
    assert df.loc[df["customer_id"] == 10, "total"].iloc[0] == 20.0


def test_select_and_rename():
    sel = get("select")
    assert sel.infer([ORDERS], {"columns": ["order_id", "amount"]}).names == ["order_id", "amount"]
    ren = get("rename")
    assert ren.infer([ORDERS], {"map": {"amount": "value"}}).has("value")


def test_join_schema_merges_and_marks_nullable():
    left = Schema.of(Field("id", "long", False), Field("a", "string"))
    right = Schema.of(Field("id", "long", False), Field("b", "double"))
    out = get("join").infer([left, right], {"on": ["id"], "how": "left"})
    assert out.names == ["id", "a", "b"]
    assert out.require("b").nullable is True  # left join -> right side nullable


def test_cast_changes_type():
    out = get("cast").infer([ORDERS], {"casts": {"amount": "decimal"}})
    assert out.require("amount").dtype == "decimal"
