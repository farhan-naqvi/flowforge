"""Authoring: create data products and write ODCS contracts + intents.

These are the write side of the GitOps flow. The UI calls them to author a new
data product, edit its source/target contracts as structured fields (no-code),
and build the pipeline intent. Everything is written as YAML into the
``covenant-contracts`` tree so the change lands through a pull request — the tool
never mutates a running system.
"""

from __future__ import annotations

import os
import re
import tempfile
from typing import List

import yaml

from covenant_transforms.schema import DTYPES

from . import gitops

SLUG_RE = re.compile(r"^[a-z][a-z0-9_]*(/[a-z][a-z0-9_]*)+$")
IDENT_RE = re.compile(r"^[A-Za-z][A-Za-z0-9_]*$")

# Covenant dtype -> ODCS logicalType (round-trips through odcs._LOGICAL).
_DTYPE_TO_LOGICAL = {
    "int": "integer", "long": "integer", "double": "number", "decimal": "decimal",
    "string": "string", "bool": "boolean", "date": "date", "timestamp": "timestamp",
}


class AuthoringError(ValueError):
    pass


def validate_slug(slug: str) -> None:
    if not SLUG_RE.match(slug):
        raise AuthoringError(
            f"invalid slug {slug!r}: use 'domain/data_product' with lowercase "
            "letters, digits, underscores (at least two segments)"
        )


def _validate_fields(fields: List[dict]) -> None:
    if not fields:
        raise AuthoringError("a contract needs at least one field")
    seen = set()
    for f in fields:
        name = f.get("name", "")
        if not IDENT_RE.match(name):
            raise AuthoringError(f"invalid field name {name!r} (letters, digits, underscore; letter first)")
        if name in seen:
            raise AuthoringError(f"duplicate field name {name!r}")
        seen.add(name)
        if f.get("type") not in DTYPES:
            raise AuthoringError(f"field {name!r}: unknown type {f.get('type')!r}; allowed: {', '.join(DTYPES)}")


def schema_to_odcs(contract_id: str, table_name: str, fields: List[dict]) -> dict:
    """Serialize structured fields to an ODCS v3 contract document."""
    _validate_fields(fields)
    props = []
    for f in fields:
        prop = {
            "name": f["name"],
            "logicalType": _DTYPE_TO_LOGICAL[f["type"]],
            "required": bool(f.get("nullable", True) is False),
        }
        if f.get("primary_key"):
            prop["primaryKey"] = True
        props.append(prop)
    return {
        "apiVersion": "v3.0.0",
        "kind": "DataContract",
        "id": contract_id,
        "name": table_name,
        "schema": [{"name": table_name, "physicalType": "table", "properties": props}],
    }


def contract_rel_path(slug: str, kind: str) -> str:
    if kind not in ("source", "target"):
        raise AuthoringError("kind must be 'source' or 'target'")
    return os.path.join(gitops.CONTRACTS_ROOT, "contracts", slug, f"{kind}.odcs.yaml")


def intent_rel_path(slug: str) -> str:
    return os.path.join(gitops.CONTRACTS_ROOT, "intents", slug, "pipeline.intent.yaml")


def product_exists(base_dir: str, slug: str) -> bool:
    return os.path.exists(os.path.join(base_dir, intent_rel_path(slug)))


def _write_yaml(base_dir: str, rel_path: str, doc: dict) -> str:
    """Atomically write *doc* as YAML to *rel_path* under *base_dir*.

    Writes to a temp file in the same directory, then ``os.replace`` — so a
    reader never observes a half-written contract, and a crash mid-write leaves
    the previous version intact.
    """
    path = os.path.join(base_dir, rel_path)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    body = "# Authored via the Covenant workbench. Review and commit via pull request.\n"
    body += yaml.safe_dump(doc, sort_keys=False)
    fd, tmp = tempfile.mkstemp(dir=os.path.dirname(path), suffix=".tmp")
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as fh:
            fh.write(body)
            fh.flush()
            os.fsync(fh.fileno())
        os.replace(tmp, path)
    except BaseException:
        if os.path.exists(tmp):
            os.unlink(tmp)
        raise
    return rel_path


def save_contract(base_dir: str, slug: str, kind: str, fields: List[dict]) -> str:
    validate_slug(slug)
    contract_id = f"{slug.replace('/', '.')}.{kind}"
    table_name = slug.split("/")[-1] + ("_raw" if kind == "source" else "")
    doc = schema_to_odcs(contract_id, table_name, fields)
    return _write_yaml(base_dir, contract_rel_path(slug, kind), doc)


def save_intent(base_dir: str, slug: str, steps: List[dict], transforms_version: str = "0.1.0") -> str:
    validate_slug(slug)
    for s in steps:
        if "primitive" not in s:
            raise AuthoringError("each step needs a 'primitive'")
    doc = {
        "data_product": slug,
        "source_contract": contract_rel_path(slug, "source"),
        "target_contract": contract_rel_path(slug, "target"),
        "transforms_version": transforms_version,
        "steps": [{"primitive": s["primitive"], "params": s.get("params", {})} for s in steps],
    }
    return _write_yaml(base_dir, intent_rel_path(slug), doc)


def create_product(base_dir: str, slug: str) -> dict:
    """Create a new data product with skeleton contracts + an empty intent."""
    validate_slug(slug)
    if os.path.exists(os.path.join(base_dir, intent_rel_path(slug))):
        raise AuthoringError(f"data product {slug!r} already exists")

    src_fields = [
        {"name": "id", "type": "long", "nullable": False, "primary_key": True},
        {"name": "value", "type": "double", "nullable": True},
    ]
    tgt_fields = [
        {"name": "id", "type": "long", "nullable": False, "primary_key": True},
        {"name": "value", "type": "double", "nullable": True},
    ]
    save_contract(base_dir, slug, "source", src_fields)
    save_contract(base_dir, slug, "target", tgt_fields)
    save_intent(base_dir, slug, steps=[])
    return {"slug": slug, "intent_path": intent_rel_path(slug)}
