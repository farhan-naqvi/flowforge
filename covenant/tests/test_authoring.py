"""Authoring: create products, edit contracts/intents, and feed the SAME verify
and compile pipeline the CLI/server use (no parallel implementation)."""

import os

import pytest
import yaml

from covenant import authoring
from covenant.gitops import discover
from covenant.model import Intent
from covenant.odcs import load_contract
from covenant.planner import plan_from_intent
from covenant.verify import verify_plan

SOURCE_FIELDS = [
    {"name": "order_id", "type": "long", "nullable": False, "primary_key": True},
    {"name": "customer_id", "type": "long", "nullable": False},
    {"name": "amount", "type": "double", "nullable": False},
    {"name": "status", "type": "string", "nullable": False},
    {"name": "created_at", "type": "timestamp", "nullable": False},
]
TARGET_FIELDS = [
    {"name": "order_date", "type": "date", "nullable": False, "primary_key": True},
    {"name": "total_amount", "type": "double", "nullable": False},
    {"name": "order_count", "type": "long", "nullable": False},
]
STEPS = [
    {"primitive": "filter", "params": {"predicate": {"op": "=", "args": [{"col": "status"}, {"lit": "completed"}]}}},
    {"primitive": "derive", "params": {"column": "order_date",
                                       "expr": {"fn": "date_trunc", "args": [{"lit": "day"}, {"col": "created_at"}], "type": "date"}}},
    {"primitive": "aggregate", "params": {"group_by": ["order_date"],
                                          "aggs": [{"name": "total_amount", "fn": "sum", "col": "amount"},
                                                   {"name": "order_count", "fn": "count", "col": "order_id"}]}},
]


def test_create_product_writes_files_and_is_discoverable(tmp_path):
    base = str(tmp_path)
    authoring.create_product(base, "sales/orders")
    assert os.path.isfile(os.path.join(base, "covenant-contracts/contracts/sales/orders/source.odcs.yaml"))
    assert os.path.isfile(os.path.join(base, "covenant-contracts/contracts/sales/orders/target.odcs.yaml"))
    assert os.path.isfile(os.path.join(base, "covenant-contracts/intents/sales/orders/pipeline.intent.yaml"))
    assert [p.slug for p in discover(base)] == ["sales/orders"]


def test_create_refuses_duplicate(tmp_path):
    base = str(tmp_path)
    authoring.create_product(base, "a/b")
    with pytest.raises(authoring.AuthoringError):
        authoring.create_product(base, "a/b")


@pytest.mark.parametrize("slug", ["nodomain", "Bad/Slug", "a/b/../c", ""])
def test_invalid_slug_rejected(tmp_path, slug):
    with pytest.raises(authoring.AuthoringError):
        authoring.create_product(str(tmp_path), slug)


def test_contract_roundtrips_types_nullability_pk(tmp_path):
    base = str(tmp_path)
    authoring.create_product(base, "sales/orders")
    authoring.save_contract(base, "sales/orders", "source", SOURCE_FIELDS)
    c = load_contract(os.path.join(base, "covenant-contracts/contracts/sales/orders/source.odcs.yaml"))
    assert c.schema.require("order_id").dtype == "long"
    assert c.schema.require("order_id").nullable is False
    assert c.schema.require("amount").dtype == "double"
    assert c.primary_key == ["order_id"]


@pytest.mark.parametrize("fields,msg", [
    ([{"name": "x", "type": "nope", "nullable": True}], "unknown type"),
    ([{"name": "2bad", "type": "string", "nullable": True}], "invalid field name"),
    ([{"name": "a", "type": "string"}, {"name": "a", "type": "long"}], "duplicate"),
    ([], "at least one field"),
])
def test_invalid_contract_rejected_with_message(tmp_path, fields, msg):
    with pytest.raises(authoring.AuthoringError) as exc:
        authoring.save_contract(str(tmp_path), "a/b", "source", fields)
    assert msg in str(exc.value)


def test_invalid_save_does_not_overwrite_existing(tmp_path):
    base = str(tmp_path)
    authoring.create_product(base, "a/b")
    authoring.save_contract(base, "a/b", "source", SOURCE_FIELDS)
    path = os.path.join(base, "covenant-contracts/contracts/a/b/source.odcs.yaml")
    before = open(path).read()
    with pytest.raises(authoring.AuthoringError):
        authoring.save_contract(base, "a/b", "source", [{"name": "bad name", "type": "string"}])
    assert open(path).read() == before          # unchanged
    # and no leftover temp files in the directory
    assert not [f for f in os.listdir(os.path.dirname(path)) if f.endswith(".tmp")]


def test_authored_product_feeds_existing_verify_pipeline(tmp_path):
    """Round-trip: author everything, then plan + verify via the SAME functions
    the CLI/server call — no parallel path."""
    base = str(tmp_path)
    authoring.create_product(base, "sales/daily_orders")
    authoring.save_contract(base, "sales/daily_orders", "source", SOURCE_FIELDS)
    authoring.save_contract(base, "sales/daily_orders", "target", TARGET_FIELDS)
    authoring.save_intent(base, "sales/daily_orders", STEPS)

    intent = Intent.load(os.path.join(base, "covenant-contracts/intents/sales/daily_orders/pipeline.intent.yaml"))
    plan = plan_from_intent(intent, base_dir=base)
    assert plan.conformance.ok, plan.conformance.problems

    target = load_contract(os.path.join(base, plan.target_contract))
    verdict = verify_plan(plan, target, base_dir=base)
    assert verdict.ok, (verdict.schema_problems, verdict.quality_problems)
    assert verdict.row_count > 0


def test_persistence_survives_reload(tmp_path):
    base = str(tmp_path)
    authoring.create_product(base, "a/b")
    authoring.save_contract(base, "a/b", "target", TARGET_FIELDS)
    # simulate a fresh process: re-read from disk only
    doc = yaml.safe_load(open(os.path.join(base, "covenant-contracts/contracts/a/b/target.odcs.yaml")))
    names = [p["name"] for p in doc["schema"][0]["properties"]]
    assert names == ["order_date", "total_amount", "order_count"]
