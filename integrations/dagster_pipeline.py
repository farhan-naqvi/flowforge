try:
    from dagster import job, op
except Exception:
    def op(func=None, **kwargs):
        if func is None:
            def wrap(f):
                return f
            return wrap
        return func
    def job(func=None, **kwargs):
        if func is None:
            def wrap(f):
                return f
            return wrap
        return func
import time
import requests
import uuid
from datetime import datetime


@op
def extract_op():
    time.sleep(1)
    r = {"rows": 15234, "size_mb": 234}
    # post task log if observability available
    try:
        run_id = getattr(extract_op, 'run_id', str(uuid.uuid4()))
        requests.post('http://localhost:8000/tasks', json={"run_id": run_id, "task_name": "extract", "status": "completed", "started_at": datetime.utcnow().isoformat(), "finished_at": datetime.utcnow().isoformat(), "logs": "extract done"})
    except Exception:
        pass
    return r


@op
def transform_op(data):
    time.sleep(2)
    data["rows"] = int(data["rows"] * 0.98)
    try:
        run_id = getattr(transform_op, 'run_id', None)
        requests.post('http://localhost:8000/tasks', json={"run_id": run_id or str(uuid.uuid4()), "task_name": "transform", "status": "completed", "started_at": datetime.utcnow().isoformat(), "finished_at": datetime.utcnow().isoformat(), "logs": "transform done"})
    except Exception:
        pass
    return {"rows": data["rows"], "cleaned": True}


@op
def load_op(data):
    time.sleep(1)
    try:
        run_id = getattr(load_op, 'run_id', None)
        requests.post('http://localhost:8000/tasks', json={"run_id": run_id or str(uuid.uuid4()), "task_name": "load", "status": "completed", "started_at": datetime.utcnow().isoformat(), "finished_at": datetime.utcnow().isoformat(), "logs": "load done"})
    except Exception:
        pass
    return {"status": "loaded", "rows": data["rows"]}


@job
def etl_job():
    # create run record
    run_id = str(uuid.uuid4())
    extract_op.run_id = run_id
    transform_op.run_id = run_id
    load_op.run_id = run_id
    try:
        requests.post('http://localhost:8000/runs', json={"id": run_id, "pipeline_id": "etl-pipeline", "status": "running", "started_at": datetime.utcnow().isoformat()})
    except Exception:
        pass
    data = extract_op()
    t = transform_op(data)
    load_op(t)
    try:
        requests.post('http://localhost:8000/runs', json={"id": run_id, "pipeline_id": "etl-pipeline", "status": "completed", "finished_at": datetime.utcnow().isoformat()})
    except Exception:
        pass


if __name__ == "__main__":
    # If dagster isn't available, call job function directly
    try:
        result = etl_job()
        print(result)
    except Exception:
        # fallback: run ops sequentially
        d = extract_op()
        t = transform_op(d)
        r = load_op(t)
        print(r)
