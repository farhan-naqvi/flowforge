try:
    from prefect import flow, task
    from prefect.tasks import task_input_hash
    from datetime import timedelta
except Exception:
    # Fall back to no-op decorators so the script can run without Prefect installed
    def task(func=None, **kwargs):
        if func is None:
            def wrapper(f):
                return f
            return wrapper
        return func
    def flow(func=None, **kwargs):
        if func is None:
            def wrapper(f):
                return f
            return wrapper
        return func
from datetime import timedelta
import time
import requests
import uuid
from datetime import datetime


@task(retries=2, retry_delay_seconds=5)
def extract():
    time.sleep(1)
    return {"rows": 15234, "size_mb": 234}


@task
def transform(data):
    time.sleep(2)
    # simple transformation
    data["rows"] = int(data["rows"] * 0.98)
    return {"rows": data["rows"], "cleaned": True}


@task
def load(data):
    time.sleep(1)
    return {"status": "loaded", "rows": data["rows"]}


@flow(name="etl-pipeline")
def etl_flow():
    run_id = str(uuid.uuid4())
    started = datetime.utcnow().isoformat()
    # announce run
    try:
        requests.post('http://localhost:8000/runs', json={"id": run_id, "pipeline_id": "etl-pipeline", "status": "running", "started_at": started})
    except Exception:
        pass
    e = extract()
    try:
        requests.post('http://localhost:8000/tasks', json={"run_id": run_id, "task_name": "extract", "status": "completed", "started_at": started, "finished_at": datetime.utcnow().isoformat(), "logs": "extracted"})
    except Exception:
        pass
    t = transform(e)
    try:
        requests.post('http://localhost:8000/tasks', json={"run_id": run_id, "task_name": "transform", "status": "running", "started_at": datetime.utcnow().isoformat(), "logs": "transform started"})
    except Exception:
        pass
    l = load(t)
    finished = datetime.utcnow().isoformat()
    try:
        requests.post('http://localhost:8000/tasks', json={"run_id": run_id, "task_name": "load", "status": "completed", "started_at": datetime.utcnow().isoformat(), "finished_at": finished, "logs": "loaded"})
        requests.post('http://localhost:8000/runs', json={"id": run_id, "pipeline_id": "etl-pipeline", "status": "completed", "started_at": started, "finished_at": finished})
    except Exception:
        pass
    return l


if __name__ == "__main__":
    # allow running without Prefect by calling the function directly
    try:
        print(etl_flow())
    except TypeError:
        # Prefect flow wrapper may require calling .run(), but our fallback returns a function
        print(etl_flow())
