from flowforge.ir import Edge, Handler, Metadata, PipelineSpec, RetryPolicy, Task, TaskPort


def _spec():
    spec = PipelineSpec(metadata=Metadata(name="p", version="1.0.0", owner="me"))
    spec.tasks["a"] = Task(type="Source", handler=Handler(source="p:a"))
    spec.tasks["b"] = Task(
        type="Sink",
        handler=Handler(source="p:b", env={"K": "V"}),
        retry=RetryPolicy(max_attempts=3, backoff="exponential", backoff_multiplier=2.0),
        timeout="30s",
        resources={"cpu": "1", "memory": "2Gi"},
    )
    spec.edges.append(Edge(TaskPort("a", "out"), TaskPort("b", "in")))
    return spec


def test_json_roundtrip():
    spec = _spec()
    restored = PipelineSpec.from_json(spec.to_json())
    assert restored.metadata.name == "p"
    assert set(restored.tasks) == {"a", "b"}
    assert restored.tasks["b"].retry.max_attempts == 3
    assert restored.tasks["b"].timeout == "30s"
    assert restored.tasks["b"].handler.env == {"K": "V"}
    assert len(restored.edges) == 1
    assert restored.edges[0].from_.task == "a"


def test_api_version_and_kind():
    d = _spec().to_dict()
    assert d["apiVersion"] == "flowforge.io/v1"
    assert d["kind"] == "Pipeline"


def test_upstream_downstream():
    spec = _spec()
    assert spec.upstream("b") == ["a"]
    assert spec.downstream("a") == ["b"]
