"""Fan-out/fan-in pipeline example."""

from flowforge import pipeline, task
from typing import Dict


@pipeline(
    name="fan_out_fan_in",
    version="1.0.0",
    description="Parallel processing with fan-out/fan-in pattern",
)
def fan_out_fan_in():
    """
    Pipeline that splits work across parallel tasks, then merges results.
    
    Flow:
    1. extract() - Read source data
    2. process_a() - Path A: Calculate sum
    3. process_b() - Path B: Calculate count (parallel)
    4. merge() - Combine results
    """
    pass


@task(image="python:3.11")
def extract() -> list:
    """Extract source data."""
    return list(range(1, 101))  # 1-100


@task(image="python:3.11")
def process_a(data: list) -> int:
    """Process path A: Calculate sum."""
    return sum(data)


@task(image="python:3.11")
def process_b(data: list) -> int:
    """Process path B: Calculate count."""
    return len(data)


@task(image="python:3.11")
def merge(sum_result: int, count_result: int) -> Dict:
    """Merge results from parallel paths."""
    return {
        "total": sum_result,
        "count": count_result,
        "average": sum_result / count_result,
    }


# Register tasks
fan_out_fan_in.add_task("extract", extract)
fan_out_fan_in.add_task("process_a", process_a)
fan_out_fan_in.add_task("process_b", process_b)
fan_out_fan_in.add_task("merge", merge)

# Connect tasks (fan-out from extract)
fan_out_fan_in.add_edge(extract, "result", process_a, "data")
fan_out_fan_in.add_edge(extract, "result", process_b, "data")

# Connect to merge (fan-in)
fan_out_fan_in.add_edge(process_a, "result", merge, "sum_result")
fan_out_fan_in.add_edge(process_b, "result", merge, "count_result")


if __name__ == "__main__":
    from flowforge import LocalExecutor, DAGVisualizer
    
    print("Pipeline Structure:")
    print(DAGVisualizer.visualize(fan_out_fan_in))
    print()
    
    print("Executing...")
    executor = LocalExecutor()
    results = executor.execute(fan_out_fan_in)
    print(f"Final result: {results['merge']}")
