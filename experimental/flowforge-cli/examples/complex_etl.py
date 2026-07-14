@pipeline(name="complex_etl", version="2.0.0", owner="ops_team")
def build_pipeline():
    pass


p = build_pipeline


@task(image="python:3.11", retries=2)
def extract() -> list:
    return []


@task
def transform(extract: list) -> list:
    return extract


@task(image="python:3.11")
def load(transform: list) -> bool:
    return True


p.add_task("extract", extract)
p.add_task("transform", transform)
p.add_task("load", load)

p.add_edge("extract", "result", transform, "data")
p.add_edge(transform, "result", "load", "data")
