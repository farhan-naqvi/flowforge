import os
import py_compile

import yaml

from flowforge.compilers import compile as compile_target
from flowforge.ir import Edge, Handler, Metadata, PipelineSpec, RetryPolicy, Task, TaskPort
from flowforge.parser import parse_file

EXAMPLES = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "examples")


def _spec():
    spec = PipelineSpec(metadata=Metadata(name="etl", owner="team"))
    spec.tasks["extract"] = Task(type="Source", handler=Handler(source="etl:extract"),
                                 executor_config={"argo": {"image": "python:3.11"}})
    spec.tasks["load"] = Task(type="Sink", handler=Handler(source="etl:load"),
                              retry=RetryPolicy(max_attempts=3), timeout="30s")
    spec.edges.append(Edge(TaskPort("extract", "out"), TaskPort("load", "in")))
    return spec


def test_argo_is_valid_yaml_with_deps():
    out = compile_target(_spec(), "argo")
    doc = yaml.safe_load(out)
    assert doc["kind"] == "Workflow"
    templates = {t["name"]: t for t in doc["spec"]["templates"]}
    dag = templates["etl"]["dag"]["tasks"]
    load = next(t for t in dag if t["name"] == "load")
    assert load["dependencies"] == ["extract"]
    assert templates["load"]["retryStrategy"]["limit"] == 3
    assert templates["load"]["activeDeadlineSeconds"] == 30


def test_airflow_is_valid_python(tmp_path):
    out = compile_target(_spec(), "airflow")
    p = tmp_path / "dag.py"
    p.write_text(out)
    py_compile.compile(str(p), doraise=True)
    assert "schedule=None" in out  # Airflow 3 style, not schedule_interval
    assert "retries=3" in out
    assert "execution_timeout=timedelta(seconds=30)" in out


def test_compilers_are_deterministic():
    spec = _spec()
    assert compile_target(spec, "argo") == compile_target(spec, "argo")
    assert compile_target(spec, "airflow") == compile_target(spec, "airflow")


def test_no_injection_in_output():
    # Even if validation were bypassed, emission must not allow breakout.
    spec = _spec()
    spec.metadata.owner = "a'; import os #"
    airflow = compile_target(spec, "airflow")
    # The dangerous string is present only inside a quoted/repr'd literal.
    for line in airflow.splitlines():
        if "import os" in line:
            assert line.strip().startswith("#") or "'owner'" in line or repr("a'; import os #") in line


def test_handler_source_newline_cannot_break_comment(tmp_path):
    # A newline in handler.source must not escape the generated comment line.
    spec = _spec()
    spec.tasks["extract"].handler = Handler(source="etl:extract\n    import os; os.system('id')")
    out = compile_target(spec, "airflow")
    p = tmp_path / "dag.py"
    p.write_text(out)
    py_compile.compile(str(p), doraise=True)
    assert "os.system" not in out or all(
        "os.system" not in ln or ln.lstrip().startswith("#") for ln in out.splitlines()
    )


def test_examples_compile_end_to_end():
    for name in ("simple_etl.py", "fan_out_fan_in.py"):
        spec = parse_file(os.path.join(EXAMPLES, name))
        argo = compile_target(spec, "argo")
        yaml.safe_load(argo)  # valid YAML
        assert compile_target(spec, "airflow")  # non-empty
