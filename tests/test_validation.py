from flowforge.ir import Edge, Handler, Metadata, PipelineSpec, Task, TaskPort
from flowforge.validation import validate


def _base(name="p"):
    spec = PipelineSpec(metadata=Metadata(name=name))
    spec.tasks["a"] = Task(type="Source", handler=Handler(source="p:a"))
    spec.tasks["b"] = Task(type="Sink", handler=Handler(source="p:b"))
    spec.edges.append(Edge(TaskPort("a", "out"), TaskPort("b", "in")))
    return spec


def test_valid_pipeline():
    assert validate(_base()).ok


def test_injection_name_rejected():
    spec = _base(name="etl'''; import os; os.system('id') #")
    result = validate(spec)
    assert not result.ok
    assert any(d.code == "name.unsafe" for d in result.errors)


def test_injection_task_id_rejected():
    spec = _base()
    spec.tasks["evil'; rm -rf /"] = Task(type="Sink", handler=Handler())
    result = validate(spec)
    assert any(d.code == "task.id.unsafe" for d in result.errors)


def test_cycle_detected():
    spec = _base()
    spec.edges.append(Edge(TaskPort("b", "out"), TaskPort("a", "in")))
    result = validate(spec)
    assert any(d.code == "graph.cycle" for d in result.errors)


def test_missing_edge_target():
    spec = _base()
    spec.edges.append(Edge(TaskPort("a", "out"), TaskPort("ghost", "in")))
    result = validate(spec)
    assert any(d.code == "edge.target.missing" for d in result.errors)


def test_isolated_node_warns():
    spec = _base()
    spec.tasks["lonely"] = Task(type="Transform", handler=Handler())
    result = validate(spec)
    assert result.ok  # only a warning
    assert any(d.code == "task.isolated" for d in result.warnings)


def test_empty_pipeline_rejected():
    spec = PipelineSpec(metadata=Metadata(name="empty"))
    result = validate(spec)
    assert any(d.code == "tasks.empty" for d in result.errors)
