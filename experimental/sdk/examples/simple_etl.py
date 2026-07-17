"""Simple ETL pipeline example."""

from flowforge import pipeline, task


@pipeline(
    name="simple_etl",
    version="1.0.0",
    owner="data_team",
    description="Simple Extract → Transform → Load pipeline",
)
def simple_etl():
    """
    A simple 3-task ETL pipeline.
    
    Flow:
    1. extract() - Read data from source
    2. transform() - Clean and transform data
    3. load() - Write to destination
    """
    pass


@task(image="python:3.11", description="Extract data from Kafka")
def extract() -> list:
    """Extract records from Kafka topic."""
    # In real execution, connect to Kafka
    return [
        {"id": 1, "name": "Alice", "value": 100},
        {"id": 2, "name": "Bob", "value": 200},
        {"id": 3, "name": "Charlie", "value": 300},
    ]


@task(image="python:3.11", description="Transform data")
def transform(data: list) -> list:
    """Clean and transform records."""
    # Convert names to uppercase and double values
    return [
        {
            "id": r["id"],
            "name": r["name"].upper(),
            "value": r["value"] * 2,
        }
        for r in data
    ]


@task(image="python:3.11", description="Load data to warehouse")
def load(data: list) -> None:
    """Write transformed data to warehouse."""
    print(f"Loaded {len(data)} records")


# Register tasks with pipeline
simple_etl.add_task("extract", extract)
simple_etl.add_task("transform", transform)
simple_etl.add_task("load", load)

# Connect tasks
simple_etl.add_edge(extract, "result", transform, "data")
simple_etl.add_edge(transform, "result", load, "data")


if __name__ == "__main__":
    from flowforge import LocalExecutor, DAGVisualizer, pipeline_to_ir
    
    # Visualize pipeline
    print("Pipeline Structure:")
    print(DAGVisualizer.visualize(simple_etl))
    print()
    
    # Export to IR
    spec = pipeline_to_ir(simple_etl)
    print("IR Specification:")
    print(spec.to_json())
    print()
    
    # Execute locally
    print("Executing pipeline...")
    executor = LocalExecutor()
    results = executor.execute(simple_etl)
    print(f"Results: {results}")
