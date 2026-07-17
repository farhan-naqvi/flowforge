"""GitOps layout helpers for the ``covenant-contracts`` source of truth.

The repo is split by concern; a shared ``<domain>/<data-product>`` slug links a
product's four artifacts across the trees:

    contracts/<slug>/{source,target}.odcs.yaml
    intents/<slug>/pipeline.intent.yaml
    plans/<slug>/pipeline.plan.yaml        # written by the planner, reviewed via PR
    tests/<slug>/{golden,fixtures}/

Writing a plan produces a file change intended to land through a pull request
(CODEOWNERS-gated), never a direct mutation of a running system.
"""

from __future__ import annotations

import os
from dataclasses import dataclass
from typing import List

CONTRACTS_ROOT = "covenant-contracts"
INTENTS_DIR = os.path.join(CONTRACTS_ROOT, "intents")
PLANS_DIR = os.path.join(CONTRACTS_ROOT, "plans")


@dataclass
class Product:
    slug: str
    intent_path: str


def discover(base_dir: str = ".") -> List[Product]:
    """Find every data product by its committed intent file."""
    root = os.path.join(base_dir, INTENTS_DIR)
    out: List[Product] = []
    if not os.path.isdir(root):
        return out
    for dirpath, _dirs, files in os.walk(root):
        for fn in files:
            if fn.endswith(".intent.yaml"):
                slug = os.path.relpath(dirpath, root)
                out.append(Product(slug=slug, intent_path=os.path.relpath(os.path.join(dirpath, fn), base_dir)))
    return sorted(out, key=lambda p: p.slug)


def plan_path(slug: str, base_dir: str = ".") -> str:
    return os.path.join(base_dir, PLANS_DIR, slug, "pipeline.plan.yaml")


def write_plan(slug: str, plan_yaml: str, base_dir: str = ".") -> str:
    path = plan_path(slug, base_dir)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", encoding="utf-8") as fh:
        fh.write(plan_yaml)
    return os.path.relpath(path, base_dir)
