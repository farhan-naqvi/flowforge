"""Conditional branching pipeline example."""

from flowforge import pipeline, task
from typing import Tuple, List


@pipeline(
    name="conditional_pipeline",
    version="1.0.0",
    description="Pipeline with conditional branching",
)
def conditional_pipeline():
    """
    Pipeline that branches based on data quality checks.
    
    Flow:
    1. extract() - Read data
    2. validate() - Check data quality
    3. Branch:
       - If valid: load_valid()
       - If invalid: alert()
    """
    pass


@task(image="python:3.11")
def extract() -> list:
    """Extract data."""
    return [
        {"id": 1, "name": "Alice", "age": 30},
        {"id": 2, "name": "Bob", "age": 25},
        {"id": 3, "name": None, "age": 35},  # Invalid: null name
    ]


@task(image="python:3.11")
def validate(data: list) -> Tuple[list, list]:
    """Validate data and split into good/bad records."""
    valid = [r for r in data if r.get("name") is not None]
    invalid = [r for r in data if r.get("name") is None]
    return valid, invalid


@task(image="python:3.11")
def load_valid(valid_data: list) -> None:
    """Load valid records."""
    print(f"✓ Loaded {len(valid_data)} valid records")


@task(image="python:3.11")
def alert(invalid_data: list) -> None:
    """Alert for invalid records."""
    print(f"⚠ Found {len(invalid_data)} invalid records: {invalid_data}")


# Register tasks
conditional_pipeline.add_task("extract", extract)
conditional_pipeline.add_task("validate", validate)
conditional_pipeline.add_task("load_valid", load_valid)
conditional_pipeline.add_task("alert", alert)

# Note: Current implementation doesn't support true conditional branching.
# In a full implementation, we'd need to route outputs based on predicates.
# For now, we show both paths executing.

conditional_pipeline.add_edge(extract, "result", validate, "data")
# In real conditional execution, validate would have multiple outputs
# and routes would be based on predicates


if __name__ == "__main__":
    from flowforge import LocalExecutor, DAGVisualizer
    
    print("Pipeline Structure:")
    print(DAGVisualizer.visualize(conditional_pipeline))
