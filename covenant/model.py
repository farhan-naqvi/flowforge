"""Intent and Plan models — the committed, reviewable artifacts.

- **Intent** is the structured, mostly-no-code description a user authors: which
  contracts, which transform steps and parameters. It is source-of-truth and
  lives in ``intents/<domain>/<data-product>/``.
- **Plan** is the resolved, schema-annotated pipeline the planner produces from
  an intent. It is deterministic to re-run and lives in ``plans/…``. CI compiles
  the execution artifact (Argo) from the plan; the artifact itself is not committed.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import List, Optional

import yaml

from covenant_transforms.schema import Schema


@dataclass
class Step:
    primitive: str
    params: dict = field(default_factory=dict)
    output_schema: Optional[Schema] = None  # filled by the planner

    def to_dict(self, include_schema: bool = True) -> dict:
        d = {"primitive": self.primitive, "params": _drop_private(self.params)}
        if include_schema and self.output_schema is not None:
            d["output_schema"] = self.output_schema.to_dicts()
        return d


@dataclass
class Intent:
    data_product: str
    source_contract: str
    target_contract: str
    steps: List[Step]
    transforms_version: str = "0.1.0"

    @classmethod
    def load(cls, path: str) -> "Intent":
        with open(path, "r", encoding="utf-8") as fh:
            doc = yaml.safe_load(fh)
        return cls.from_dict(doc)

    @classmethod
    def from_dict(cls, doc: dict) -> "Intent":
        return cls(
            data_product=doc.get("data_product", ""),
            source_contract=doc["source_contract"],
            target_contract=doc["target_contract"],
            transforms_version=str(doc.get("transforms_version", "0.1.0")),
            steps=[Step(s["primitive"], s.get("params", {})) for s in doc.get("steps", [])],
        )


@dataclass
class Conformance:
    ok: bool
    problems: List[str] = field(default_factory=list)

    def to_dict(self) -> dict:
        return {"ok": self.ok, "problems": self.problems}


@dataclass
class Plan:
    data_product: str
    source_contract: str
    target_contract: str
    transforms_version: str
    steps: List[Step]
    conformance: Conformance

    def to_dict(self) -> dict:
        return {
            "data_product": self.data_product,
            "source_contract": self.source_contract,
            "target_contract": self.target_contract,
            "transforms_version": self.transforms_version,
            "steps": [s.to_dict() for s in self.steps],
            "conformance": self.conformance.to_dict(),
        }

    def to_yaml(self) -> str:
        return yaml.safe_dump(self.to_dict(), sort_keys=False)

    @property
    def output_schema(self) -> Optional[Schema]:
        return self.steps[-1].output_schema if self.steps else None


def _drop_private(params: dict) -> dict:
    return {k: v for k, v in params.items() if not k.startswith("_")}
