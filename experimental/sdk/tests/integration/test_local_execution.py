"""Tests for local executor."""

import pytest
from flowforge.core.pipeline import pipeline
from flowforge.core.task import task
from flowforge.executor.local import LocalExecutor


def test_local_executor_simple():
    """Test local execution of simple pipeline."""
    @pipeline()
    def p():
        pass
    
    @task()
    def source() -> int:
        return 42
    
    p.add_task("source", source)
    
    executor = LocalExecutor()
    results = executor.execute(p)
    
    assert "source" in results
    assert results["source"] == 42


def test_local_executor_chain():
    """Test local execution with chained tasks."""
    @pipeline()
    def p():
        pass
    
    @task()
    def t1() -> int:
        return 10
    
    @task()
    def t2(value: int) -> int:
        return value * 2
    
    @task()
    def t3(value: int) -> int:
        return value + 5
    
    p.add_task("t1", t1)
    p.add_task("t2", t2)
    p.add_task("t3", t3)
    
    p.add_edge(t1, "result", t2, "value")
    p.add_edge(t2, "result", t3, "value")
    
    executor = LocalExecutor()
    results = executor.execute(p)
    
    # t1: 10, t2: 10*2=20, t3: 20+5=25
    assert results["t1"] == 10
    assert results["t2"] == 20
    assert results["t3"] == 25


def test_local_executor_validation():
    """Test that invalid pipeline raises error."""
    @pipeline()
    def p():
        pass
    
    executor = LocalExecutor()
    
    with pytest.raises(ValueError):
        executor.execute(p)  # Empty pipeline


def test_local_executor_with_cache():
    """Test local executor with caching."""
    @pipeline()
    def p():
        pass
    
    @task()
    def source() -> list:
        return [1, 2, 3]
    
    p.add_task("source", source)
    
    executor = LocalExecutor(cache_dir="./.test_cache")
    results = executor.execute(p)
    
    assert results["source"] == [1, 2, 3]


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
