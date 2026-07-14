@pipeline(name="etl", version="1.0.0", owner="data_team")
def my_pipeline():
    pass


@task(image="python:3.11", timeout="3600s", retries=3)
def extract() -> list:
    return []


@task(image="python:3.11")
def transform(data: list) -> list:
    return data


my_pipeline.add_task("extract", extract)
my_pipeline.add_task("transform", transform)
my_pipeline.add_edge(extract, "result", transform, "data")
