"""Integration tests for end-to-end pipeline usage."""

import pytest
from flowforge.core.pipeline import pipeline
from flowforge.core.task import task
from flowforge.compiler.ir_exporter import pipeline_to_ir
from flowforge.executor.local import LocalExecutor
from flowforge.graph.visualizer import DAGVisualizer


def test_simple_etl_pipeline():
    """Test end-to-end simple ETL pipeline."""
    @pipeline(name="simple_etl", version="1.0.0")
    def etl():
        pass
    
    @task(image="python:3.11")
    def extract() -> list:
        return [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
    
    @task(image="python:3.11")
    def transform(data: list) -> list:
        return [{"id": r["id"], "name": r["name"].upper()} for r in data]
    
    @task(image="python:3.11")
    def load(data: list) -> None:
        pass
    
    etl.add_task("extract", extract)
    etl.add_task("transform", transform)
    etl.add_task("load", load)
    
    etl.add_edge(extract, "result", transform, "data")
    etl.add_edge(transform, "result", load, "data")
    
    # Validate
    errors = etl.validate()
    assert len(errors) == 0
    
    # Export to IR
    spec = pipeline_to_ir(etl)
    assert spec.metadata.name == "simple_etl"
    assert len(spec.tasks) == 3
    assert len(spec.edges) == 2
    
    # Execute locally
    executor = LocalExecutor()
    results = executor.execute(etl)
    assert "extract" in etl.tasks
    assert len(results) == 3


def test_pipeline_ir_export():
    """Test pipeline export to IR."""
    @pipeline(name="test_pipeline", owner="data_team")
    def my_pipeline():
        pass
    
    @task(image="test:v1")
    def task1() -> list:
        return []
    
    @task(image="test:v2")
    def task2(data: list) -> list:
        return data
    
    my_pipeline.add_task("task1", task1)
    my_pipeline.add_task("task2", task2)
    my_pipeline.add_edge(task1, "result", task2, "data")
    
    spec = pipeline_to_ir(my_pipeline)
    
    assert spec.metadata.name == "test_pipeline"
    assert spec.metadata.owner == "data_team"
    assert len(spec.tasks) == 2
    assert spec.tasks["task1"].config.image == "test:v1"


def test_pipeline_visualization():
    """Test pipeline visualization."""
    @pipeline(name="test_pipeline")
    def my_pipeline():
        pass
    
    @task()
    def task1() -> list:
        return []
    
    @task()
    def task2(data: list) -> list:
        return data
    
    my_pipeline.add_task("task1", task1)
    my_pipeline.add_task("task2", task2)
    my_pipeline.add_edge(task1, "result", task2, "data")
    
    viz = DAGVisualizer.visualize(my_pipeline)
    assert "test_pipeline" in viz
    assert "task1" in viz
    assert "task2" in viz


def test_fan_out_fan_in_pipeline():
    """Test fan-out/fan-in pipeline pattern."""
    @pipeline(name="fan_out_fan_in")
    def my_pipeline():
        pass
    
    @task()
    def source() -> list:
        return [1, 2, 3]
    
    @task()
    def process_a(data: list) -> int:
        return sum(data)
    
    @task()
    def process_b(data: list) -> int:
        return len(data)
    
    @task()
    def merge(a: int, b: int) -> dict:
        return {"sum": a, "count": b}
    
    my_pipeline.add_task("source", source)
    my_pipeline.add_task("process_a", process_a)
    my_pipeline.add_task("process_b", process_b)
    my_pipeline.add_task("merge", merge)
    
    my_pipeline.add_edge(source, "result", process_a, "data")
    my_pipeline.add_edge(source, "result", process_b, "data")
    my_pipeline.add_edge(process_a, "result", merge, "a")
    my_pipeline.add_edge(process_b, "result", merge, "b")
    
    errors = my_pipeline.validate()
    assert len(errors) == 0
    
    spec = pipeline_to_ir(my_pipeline)
    assert len(spec.edges) == 4


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
