"""A minimal, runnable FlowForge pipeline: extract -> transform -> load.

    ff validate   examples/simple_etl.py
    ff capability examples/simple_etl.py
    ff compile    examples/simple_etl.py argo
    ff run        examples/simple_etl.py
"""

from flowforge import pipeline, task


@task(image="python:3.11")
def extract():
    """Produce some rows."""
    rows = [{"id": i, "value": i * 10} for i in range(5)]
    print(f"extracted {len(rows)} rows")
    return rows


@task(retries=3, timeout="30s")
def transform(extract):
    """Double every value."""
    out = [{**r, "value": r["value"] * 2} for r in extract]
    print(f"transformed {len(out)} rows")
    return out


@task
def load(transform):
    """Summarise the result."""
    total = sum(r["value"] for r in transform)
    print(f"loaded {len(transform)} rows, total value {total}")
    return {"rows": len(transform), "total": total}


etl = pipeline(
    "simple-etl",
    tasks=[extract, transform, load],
    version="1.0.0",
    owner="data-team",
    description="Extract, double values, summarise.",
)
