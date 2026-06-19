"""Unit tests for task decorator and Task class."""

import pytest
from flowforge.core.task import task, Task, TaskConfig, TaskInputPort, TaskOutputPort
from typing import List, Dict


def test_task_decorator_basic():
    """Test basic task decorator."""
    @task()
    def my_task():
        return "result"
    
    assert isinstance(my_task, Task)
    assert my_task.name == "my_task"
    assert len(my_task.inputs) == 0
    assert len(my_task.outputs) == 0


def test_task_with_inputs_outputs():
    """Test task with type hints for inputs/outputs."""
    @task()
    def extract(topic: str) -> List[Dict]:
        return [{"id": 1}]
    
    assert "topic" in extract.inputs
    assert "result" in extract.outputs
    assert extract.inputs["topic"].type_hint == str
    assert extract.outputs["result"].type_hint == List[Dict]


def test_task_config():
    """Test task with configuration."""
    @task(
        image="python:3.11",
        timeout="3600s",
        retries=3,
        description="Extract data",
    )
    def extract():
        pass
    
    assert extract.config.image == "python:3.11"
    assert extract.config.timeout == "3600s"
    assert extract.config.retries == 3
    assert extract.description == "Extract data"


def test_task_execution():
    """Test task execution."""
    @task()
    def my_task() -> int:
        return 42
    
    result = my_task()
    assert result == 42
    assert my_task.has_executed()
    assert my_task.get_result() == 42


def test_task_with_arguments():
    """Test task execution with arguments."""
    @task()
    def add(a: int, b: int) -> int:
        return a + b
    
    result = add(a=10, b=20)
    assert result == 30


def test_task_reset():
    """Test task reset."""
    @task()
    def my_task() -> int:
        return 42
    
    my_task()
    assert my_task.has_executed()
    
    my_task.reset()
    assert not my_task.has_executed()
    assert my_task.get_result() is None


def test_multiple_tasks():
    """Test multiple task instances."""
    @task()
    def task1() -> int:
        return 1
    
    @task()
    def task2() -> int:
        return 2
    
    assert task1.task_id != task2.task_id
    assert task1.name == "task1"
    assert task2.name == "task2"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
