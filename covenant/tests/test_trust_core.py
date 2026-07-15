import os

import pytest
from conftest import BASE_DIR

from covenant.compile import compile_argo
from covenant.model import Intent, Step
from covenant.odcs import parse_contract
from covenant.planner import DeterministicPlanner, plan_from_intent
from covenant.verify import verify_plan

EXAMPLE_INTENT = os.path.join(
    BASE_DIR, "covenant-contracts/intents/sales/daily_orders/pipeline.intent.yaml"
)


def test_example_plan_statically_conforms():
    intent = Intent.load(EXAMPLE_INTENT)
    plan = plan_from_intent(intent, base_dir=BASE_DIR)
    assert plan.conformance.ok, plan.conformance.problems
    assert [s.primitive for s in plan.steps] == ["source", "filter", "derive", "aggregate", "sink"]


def test_example_verifies_dynamically():
    intent = Intent.load(EXAMPLE_INTENT)
    plan = plan_from_intent(intent, base_dir=BASE_DIR)
    from covenant.odcs import load_contract

    target = load_contract(os.path.join(BASE_DIR, plan.target_contract))
    verdict = verify_plan(plan, target, base_dir=BASE_DIR)
    assert verdict.ok, (verdict.schema_problems, verdict.quality_problems)
    assert verdict.row_count > 0


def test_static_rejects_missing_target_column():
    # An intent that only selects amount cannot satisfy a 3-column target.
    intent = Intent.load(EXAMPLE_INTENT)
    intent.steps = [Step("select", {"columns": ["amount"]})]
    plan = plan_from_intent(intent, base_dir=BASE_DIR)
    assert not plan.conformance.ok
    assert any("order_date" in p for p in plan.conformance.problems)


def test_dynamic_catches_what_static_misses():
    """The key trust property: a plan whose *schema* satisfies a target but whose
    *data* violates a quality rule is caught only by the dynamic run."""
    source = parse_contract(
        {
            "id": "s",
            "schema": [
                {
                    "name": "orders",
                    "properties": [
                        {"name": "order_id", "logicalType": "integer", "required": True, "primaryKey": True},
                        {"name": "created_at", "logicalType": "timestamp", "required": True},
                    ],
                }
            ],
        }
    )
    # Target: order_date is the primary key (must be unique).
    target = parse_contract(
        {
            "id": "t",
            "schema": [
                {
                    "name": "daily",
                    "properties": [
                        {"name": "order_date", "logicalType": "date", "required": True, "primaryKey": True},
                    ],
                }
            ],
        }
    )
    # Plan derives order_date but does NOT aggregate -> many rows share a date.
    intent = Intent(
        data_product="x",
        source_contract="",
        target_contract="",
        steps=[
            Step("derive", {"column": "order_date",
                            "expr": {"fn": "date_trunc", "args": [{"lit": "day"}, {"col": "created_at"}], "type": "date"}}),
            Step("select", {"columns": ["order_date"]}),
        ],
    )
    plan = DeterministicPlanner().plan(intent, source, target)
    assert plan.conformance.ok  # STATIC passes: order_date is present and typed

    verdict = verify_plan(plan, target, source=source, test_data=_multi_day_rows())
    assert not verdict.ok  # DYNAMIC catches the duplicate primary key
    assert any("unique" in p for p in verdict.quality_problems)


def _multi_day_rows():
    import datetime as dt

    return {
        "order_id": [1, 2, 3, 4],
        "created_at": [dt.datetime(2026, 1, 1, 9), dt.datetime(2026, 1, 1, 18),
                       dt.datetime(2026, 1, 2, 9), dt.datetime(2026, 1, 2, 18)],
    }


def test_compile_refuses_nonconformant_plan():
    intent = Intent.load(EXAMPLE_INTENT)
    intent.steps = [Step("select", {"columns": ["amount"]})]
    plan = plan_from_intent(intent, base_dir=BASE_DIR)
    with pytest.raises(ValueError):
        compile_argo(plan, plan_path="x")


def test_compile_produces_valid_argo_yaml():
    import yaml

    intent = Intent.load(EXAMPLE_INTENT)
    plan = plan_from_intent(intent, base_dir=BASE_DIR)
    art = compile_argo(plan, plan_path="plans/sales/daily_orders/pipeline.plan.yaml")
    doc = yaml.safe_load(art)
    assert doc["kind"] == "Workflow"
    assert doc["spec"]["templates"][0]["container"]["args"][-1] == "spark"


def test_odcs_extracts_schema_and_pk():
    c = parse_contract(
        {
            "id": "c",
            "schema": [
                {
                    "name": "t",
                    "properties": [
                        {"name": "id", "logicalType": "integer", "required": True, "primaryKey": True},
                        {"name": "note", "logicalType": "string"},
                    ],
                }
            ],
        }
    )
    assert c.schema.require("id").dtype == "long"
    assert c.schema.require("id").nullable is False
    assert c.schema.require("note").nullable is True
    assert c.primary_key == ["id"]
