"""Test fixture utilities for sample pipelines."""

from flowforge import pipeline, task
from typing import List, Dict


def create_simple_pipeline():
    """Create a simple 3-task pipeline for testing."""
    @pipeline(name="test_simple")
    def p():
        pass
    
    @task()
    def t1() -> list:
        return [1, 2, 3]
    
    @task()
    def t2(data: list) -> list:
        return [x * 2 for x in data]
    
    @task()
    def t3(data: list) -> int:
        return sum(data)
    
    p.add_task("t1", t1)
    p.add_task("t2", t2)
    p.add_task("t3", t3)
    p.add_edge(t1, "result", t2, "data")
    p.add_edge(t2, "result", t3, "data")
    
    return p


def create_fan_out_pipeline():
    """Create a fan-out/fan-in pipeline for testing."""
    @pipeline(name="test_fan_out")
    def p():
        pass
    
    @task()
    def source() -> list:
        return [1, 2, 3, 4, 5]
    
    @task()
    def sum_path(data: list) -> int:
        return sum(data)
    
    @task()
    def count_path(data: list) -> int:
        return len(data)
    
    @task()
    def merge(s: int, c: int) -> dict:
        return {"sum": s, "count": c}
    
    p.add_task("source", source)
    p.add_task("sum_path", sum_path)
    p.add_task("count_path", count_path)
    p.add_task("merge", merge)
    
    p.add_edge(source, "result", sum_path, "data")
    p.add_edge(source, "result", count_path, "data")
    p.add_edge(sum_path, "result", merge, "s")
    p.add_edge(count_path, "result", merge, "c")
    
    return p


__all__ = ['create_simple_pipeline', 'create_fan_out_pipeline']
