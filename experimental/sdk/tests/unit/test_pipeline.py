"""Unit tests for pipeline decorator and Pipeline class."""

import pytest
from flowforge.core.pipeline import pipeline, Pipeline
from flowforge.core.task import task


@pipeline()
def simple_pipeline():
    """Simple 3-task pipeline."""
    pass


def test_pipeline_decorator_basic():
    """Test basic pipeline decorator."""
    assert isinstance(simple_pipeline, Pipeline)
    assert simple_pipeline.name == "simple_pipeline"


def test_pipeline_config():
    """Test pipeline with configuration."""
    @pipeline(
        name="etl_pipeline",
        version="1.0.0",
        owner="data_team",
        description="Extract Transform Load",
    )
    def my_pipeline():
        pass
    
    assert my_pipeline.config.name == "etl_pipeline"
    assert my_pipeline.config.version == "1.0.0"
    assert my_pipeline.config.owner == "data_team"
    assert "Extract Transform Load" in my_pipeline.description


def test_pipeline_with_tasks():
    """Test pipeline with tasks (manual registration)."""
    @pipeline()
    def my_pipeline():
        pass
    
    @task()
    def extract():
        return []
    
    @task()
    def load(data):
        pass
    
    my_pipeline.add_task("extract", extract)
    my_pipeline.add_task("load", load)
    
    assert len(my_pipeline.tasks) == 2
    assert "extract" in my_pipeline.tasks
    assert "load" in my_pipeline.tasks


def test_pipeline_with_edges():
    """Test pipeline with edges."""
    @pipeline()
    def my_pipeline():
        pass
    
    @task()
    def extract() -> list:
        return []
    
    @task()
    def load(data: list):
        pass
    
    my_pipeline.add_task("extract", extract)
    my_pipeline.add_task("load", load)
    my_pipeline.add_edge(extract, "result", load, "data")
    
    assert len(my_pipeline.edges) == 1
    edge = my_pipeline.edges[0]
    assert edge[0] == extract
    assert edge[1] == "result"
    assert edge[2] == load
    assert edge[3] == "data"


def test_pipeline_validation_empty():
    """Test validation of empty pipeline."""
    @pipeline()
    def my_pipeline():
        pass
    
    errors = my_pipeline.validate()
    assert len(errors) > 0
    assert "no tasks" in str(errors)


def test_pipeline_validation_valid():
    """Test validation of valid pipeline."""
    @pipeline()
    def my_pipeline():
        pass
    
    @task()
    def task1() -> list:
        return []
    
    my_pipeline.add_task("task1", task1)
    
    errors = my_pipeline.validate()
    assert len(errors) == 0


def test_source_and_sink_tasks():
    """Test identification of source and sink tasks."""
    @pipeline()
    def my_pipeline():
        pass
    
    @task()
    def extract() -> list:
        return []
    
    @task()
    def transform(data: list) -> list:
        return data
    
    @task()
    def load(data: list):
        pass
    
    my_pipeline.add_task("extract", extract)
    my_pipeline.add_task("transform", transform)
    my_pipeline.add_task("load", load)
    
    my_pipeline.add_edge(extract, "result", transform, "data")
    my_pipeline.add_edge(transform, "result", load, "data")
    
    sources = my_pipeline.get_source_tasks()
    sinks = my_pipeline.get_sink_tasks()
    
    assert len(sources) == 1
    assert sources[0] == extract
    assert len(sinks) == 1
    assert sinks[0] == load


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
