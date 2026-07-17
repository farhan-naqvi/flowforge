"""Planner: intent -> schema-annotated plan + conformance verdict.

The planner threads a schema through the pipeline using each primitive's static
``infer`` (no data moves) and checks the final schema against the target
contract. This is the plan-time half of the trust story: a plan that cannot
produce the target schema is rejected before any execution.

``Planner`` is an interface. ``DeterministicPlanner`` composes the steps the user
authored (v1). An LLM-backed planner is a drop-in that *proposes* steps — but it
is validated by the exact same conformance check here, so an agent can never
emit an unverified plan.
"""

from __future__ import annotations

from typing import List

from covenant_transforms import get as get_primitive
from covenant_transforms.schema import Schema

from .model import Conformance, Intent, Plan, Step
from .odcs import Contract, load_contract


class PlanError(ValueError):
    pass


class Planner:
    def plan(self, intent: Intent, source: Contract, target: Contract) -> Plan:
        raise NotImplementedError


class DeterministicPlanner(Planner):
    """Compose the user-authored steps, threading schemas and checking the target.

    v1 assumes a linear chain: each step consumes the previous step's output as
    its single input. Multi-input primitives (join/union) are a plan-graph
    extension tracked in the roadmap.
    """

    def plan(self, intent: Intent, source: Contract, target: Contract) -> Plan:
        steps: List[Step] = []

        # 1. Source step: schema comes from the source contract.
        src = Step("source", {"name": "source", "schema": source.schema.to_dicts()})
        src.output_schema = source.schema
        steps.append(src)

        # 2. User steps: infer output schema for each, in order.
        current: Schema = source.schema
        for i, s in enumerate(intent.steps):
            prim = get_primitive(s.primitive)
            if prim.arity == 2:
                raise PlanError(
                    f"step {i} ('{s.primitive}'): multi-input primitives are not "
                    "supported in a linear v1 intent"
                )
            try:
                current = prim.infer([current], s.params)
            except Exception as exc:  # noqa: BLE001 - surface as a planning error
                raise PlanError(f"step {i} ('{s.primitive}'): {exc}") from exc
            step = Step(s.primitive, s.params)
            step.output_schema = current
            steps.append(step)

        # 3. Sink step (terminal).
        sink = Step("sink", {})
        sink.output_schema = current
        steps.append(sink)

        # 4. Conformance: does the final schema satisfy the target contract?
        problems = current.conformance(target.schema)
        conformance = Conformance(ok=not problems, problems=problems)

        return Plan(
            data_product=intent.data_product,
            source_contract=intent.source_contract,
            target_contract=intent.target_contract,
            transforms_version=intent.transforms_version,
            steps=steps,
            conformance=conformance,
        )


def plan_from_intent(intent: Intent, base_dir: str = ".", planner: Planner | None = None) -> Plan:
    """Load the two contracts referenced by *intent* and produce a plan."""
    import os

    source = load_contract(os.path.join(base_dir, intent.source_contract))
    target = load_contract(os.path.join(base_dir, intent.target_contract))
    return (planner or DeterministicPlanner()).plan(intent, source, target)
