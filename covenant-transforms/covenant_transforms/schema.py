"""The schema currency used for plan-time inference.

A ``Schema`` is an ordered list of typed, nullable fields. Every primitive maps
input schemas to an output schema *without running* (``Primitive.infer``), so
Covenant can check a plan against a target contract before any data moves. The
same ``Schema`` maps to/from an Ibis schema for actual execution.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import List, Optional, Tuple

# Canonical logical types. Kept small and backend-neutral; mapped to Ibis on
# execution and to ODCS logical types on contract load.
DTYPES = ("int", "long", "double", "decimal", "string", "bool", "date", "timestamp")

_IBIS_TO_DTYPE = {
    "int8": "int", "int16": "int", "int32": "int", "int64": "long",
    "uint8": "int", "uint16": "int", "uint32": "int", "uint64": "long",
    "float32": "double", "float64": "double",
    "decimal": "decimal", "string": "string", "boolean": "bool",
    "date": "date", "timestamp": "timestamp",
}
_DTYPE_TO_IBIS = {
    "int": "int32", "long": "int64", "double": "float64", "decimal": "decimal",
    "string": "string", "bool": "boolean", "date": "date", "timestamp": "timestamp",
}


@dataclass(frozen=True)
class Field:
    name: str
    dtype: str
    nullable: bool = True

    def __post_init__(self) -> None:
        if self.dtype not in DTYPES:
            raise ValueError(f"unknown dtype {self.dtype!r}; known: {', '.join(DTYPES)}")

    def to_dict(self) -> dict:
        return {"name": self.name, "type": self.dtype, "nullable": self.nullable}


@dataclass(frozen=True)
class Schema:
    fields: Tuple[Field, ...]

    @classmethod
    def of(cls, *fields: Field) -> "Schema":
        return cls(tuple(fields))

    @classmethod
    def from_dicts(cls, items: List[dict]) -> "Schema":
        return cls(tuple(Field(i["name"], i["type"], i.get("nullable", True)) for i in items))

    # -- lookups ---------------------------------------------------------
    @property
    def names(self) -> List[str]:
        return [f.name for f in self.fields]

    def get(self, name: str) -> Optional[Field]:
        for f in self.fields:
            if f.name == name:
                return f
        return None

    def require(self, name: str) -> Field:
        f = self.get(name)
        if f is None:
            raise KeyError(f"column {name!r} not in schema ({', '.join(self.names)})")
        return f

    def has(self, name: str) -> bool:
        return self.get(name) is not None

    # -- transforms (return new schemas) --------------------------------
    def select(self, names: List[str]) -> "Schema":
        return Schema(tuple(self.require(n) for n in names))

    def drop(self, names: List[str]) -> "Schema":
        drop = set(names)
        return Schema(tuple(f for f in self.fields if f.name not in drop))

    def rename(self, mapping: dict) -> "Schema":
        return Schema(tuple(Field(mapping.get(f.name, f.name), f.dtype, f.nullable) for f in self.fields))

    def with_field(self, field: Field) -> "Schema":
        replaced = [field if f.name == field.name else f for f in self.fields]
        if field.name not in self.names:
            replaced.append(field)
        return Schema(tuple(replaced))

    # -- interop / conformance ------------------------------------------
    def to_ibis(self) -> dict:
        return {f.name: _DTYPE_TO_IBIS[f.dtype] for f in self.fields}

    @classmethod
    def from_ibis(cls, ibis_schema) -> "Schema":
        fields = []
        for name, dtype in ibis_schema.items():
            base = str(dtype).split("(")[0].lower().rstrip("?")
            fields.append(Field(name, _IBIS_TO_DTYPE.get(base, "string"), bool(getattr(dtype, "nullable", True))))
        return cls(tuple(fields))

    def to_dicts(self) -> List[dict]:
        return [f.to_dict() for f in self.fields]

    def conformance(self, target: "Schema") -> List[str]:
        """Return human-readable reasons this schema does NOT satisfy *target*.

        Empty list == conforms. A source may have extra columns; the target's
        columns must all be present with a compatible type, and a non-nullable
        target column may not be fed by a nullable producer.
        """
        problems: List[str] = []
        for tf in target.fields:
            sf = self.get(tf.name)
            if sf is None:
                problems.append(f"missing target column '{tf.name}' ({tf.dtype})")
                continue
            if not _type_compatible(sf.dtype, tf.dtype):
                problems.append(f"column '{tf.name}': produces {sf.dtype}, target requires {tf.dtype}")
            if not tf.nullable and sf.nullable:
                problems.append(f"column '{tf.name}': target is non-nullable but producer may emit nulls")
        return problems


# Widening compatibility: producing int where a long is required is fine, etc.
_WIDENS = {
    "int": {"int", "long", "double", "decimal"},
    "long": {"long", "double", "decimal"},
    "double": {"double"},
    "decimal": {"decimal", "double"},
    "string": {"string"},
    "bool": {"bool"},
    "date": {"date", "timestamp"},
    "timestamp": {"timestamp"},
}


def _type_compatible(produced: str, required: str) -> bool:
    return required in _WIDENS.get(produced, {produced})
