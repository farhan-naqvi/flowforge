import pytest

from flowforge.parser import ParseError, parse_module

SIMPLE = """
from flowforge import pipeline, task

@task(image="python:3.11")
def extract():
    return []

@task(retries=3, timeout="30s")
def transform(extract):
    return extract

@task
def load(transform):
    return True

p = pipeline("etl", tasks=[extract, transform, load], owner="team", version="1.0.0")
"""


def test_parses_tasks_and_edges():
    spec = parse_module(SIMPLE)
    assert spec.metadata.name == "etl"
    assert spec.metadata.owner == "team"
    assert set(spec.tasks) == {"extract", "transform", "load"}
    pairs = {(e.from_.task, e.to.task) for e in spec.edges}
    assert pairs == {("extract", "transform"), ("transform", "load")}


def test_task_types_inferred():
    spec = parse_module(SIMPLE)
    assert spec.tasks["extract"].type == "Source"
    assert spec.tasks["transform"].type == "Transform"
    assert spec.tasks["load"].type == "Sink"


def test_task_options_mapped():
    spec = parse_module(SIMPLE)
    assert spec.tasks["transform"].retry.max_attempts == 3
    assert spec.tasks["transform"].timeout == "30s"
    assert spec.tasks["extract"].executor_config["argo"]["image"] == "python:3.11"


def test_static_parse_matches_runtime_to_ir():
    # The AST parser must agree with authoring.Pipeline.to_ir on the same source.
    import types

    mod = types.ModuleType("m")
    exec(compile(SIMPLE, "m", "exec"), mod.__dict__)
    live = mod.p.to_ir()
    static = parse_module(SIMPLE)
    assert live.to_dict()["tasks"].keys() == static.to_dict()["tasks"].keys()
    assert {(e.from_.task, e.to.task) for e in live.edges} == {
        (e.from_.task, e.to.task) for e in static.edges
    }


def test_no_tasks_raises():
    with pytest.raises(ParseError):
        parse_module("x = 1\n")
