# covenant-transforms

The versioned **primitive library** that is Covenant's core asset. A primitive is
a declarative, typed transform that can (1) **infer its output schema statically**
and (2) **lower to Ibis** — so the same transform runs on DuckDB locally (fast,
cheap verify) and Spark at full scale, and Covenant can check a plan against a
data contract *before any data moves*.

> This package is designed to live in its **own semver-versioned repository**;
> contracts pin a version for reproducibility. It is vendored inside the Covenant
> repo during early development and can be extracted verbatim.

## Primitive interface

```python
class Primitive:
    id: str
    version: str
    arity: int                      # input tables (2 for join/union)
    def infer(inputs: list[Schema], params: dict) -> Schema:   ...   # static
    def lower(inputs: list[ibis.Table], params: dict) -> ibis.Table: ...
```

## v1 primitives (S3 Delta → Delta)

`source`, `sink`, `select`, `drop`, `rename`, `cast`, `filter`, `derive`
(with_column over a constrained expression grammar), `dedup`, `aggregate`,
`union`, `join`.

Roadmap: `scd2`, `window`, and Delta-native incremental read/write wrappers.

## Expression grammar

`filter`/`derive` use a constrained, analyzable grammar (not arbitrary Python),
so result types are inferable and lowering is safe:

```yaml
{op: "*", args: [{col: amount}, {lit: 2}]}
{fn: date_trunc, args: [{lit: day}, {col: created_at}], type: date}
```

## Test

```bash
pip install -e ".[dev]"
pytest
```
