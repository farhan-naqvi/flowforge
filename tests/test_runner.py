from flowforge.authoring import pipeline, task
from flowforge.runner import run_pipeline
from flowforge.store import Store


def _pipeline():
    @task
    def source():
        print("producing")
        return [1, 2, 3]

    @task
    def double(source):
        return [x * 2 for x in source]

    @task
    def total(double):
        return sum(double)

    return pipeline("run-test", tasks=[source, double, total])


def test_real_execution_passes_outputs():
    result = run_pipeline(_pipeline())
    assert result.status == "succeeded"
    by = {t.task: t for t in result.tasks}
    assert by["total"].result_repr == "12"
    assert "producing" in by["source"].logs


def test_failure_propagates_and_skips_downstream():
    @task
    def ok():
        return 1

    @task
    def boom(ok):
        raise RuntimeError("kaboom")

    @task
    def after(boom):
        return boom

    result = run_pipeline(pipeline("failing", tasks=[ok, boom, after]))
    assert result.status == "failed"
    by = {t.task: t for t in result.tasks}
    assert by["ok"].status == "succeeded"
    assert by["boom"].status == "failed"
    assert "kaboom" in by["boom"].error
    assert by["after"].status == "skipped"


def test_invalid_pipeline_not_executed():
    @task
    def a(b):
        return b

    @task
    def b(a):
        return a

    result = run_pipeline(pipeline("cyclic", tasks=[a, b]))
    assert result.status == "invalid"
    assert result.tasks == []


def test_events_persisted():
    store = Store(":memory:")
    result = run_pipeline(_pipeline(), store=store)
    run = store.get_run(result.run_id)
    assert run["status"] == "succeeded"
    assert len(run["tasks"]) == 3
    assert run["mode"] == "local"
