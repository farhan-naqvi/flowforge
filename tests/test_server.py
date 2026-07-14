import json
import os
import threading
import urllib.request
from http.server import ThreadingHTTPServer

import pytest

from flowforge.server import Handler, _Context
from flowforge.store import Store

EXAMPLES = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "examples")


@pytest.fixture()
def server():
    Handler.ctx = _Context(EXAMPLES, Store(":memory:"))
    httpd = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    port = httpd.server_address[1]
    t = threading.Thread(target=httpd.serve_forever, daemon=True)
    t.start()
    yield f"http://127.0.0.1:{port}"
    httpd.shutdown()


def _get(url, host=None):
    req = urllib.request.Request(url)
    if host:
        req.add_header("Host", host)
    with urllib.request.urlopen(req) as r:
        return r.status, json.loads(r.read())


def _post(url, payload):
    data = json.dumps(payload).encode()
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req) as r:
        return r.status, json.loads(r.read())


def test_health(server):
    status, body = _get(server + "/api/health")
    assert status == 200 and body["status"] == "ok"


def test_pipelines_lists_examples(server):
    _, body = _get(server + "/api/pipelines")
    names = {p["name"] for p in body["pipelines"]}
    assert "simple-etl" in names


def test_path_traversal_blocked(server):
    with pytest.raises(urllib.error.HTTPError) as exc:
        _get(server + "/api/pipeline?file=../flowforge/cli.py")
    assert exc.value.code == 404


def test_dns_rebinding_host_rejected(server):
    with pytest.raises(urllib.error.HTTPError) as exc:
        _get(server + "/api/health", host="evil.example.com")
    assert exc.value.code == 403


def test_pipeline_detail_has_validation_and_capability(server):
    _, body = _get(server + "/api/pipeline?file=simple_etl.py")
    assert body["validation"]["ok"] is True
    assert "argo" in body["capability"] and "airflow" in body["capability"]


def test_run_executes_and_persists(server):
    status, body = _post(server + "/api/run", {"file": "simple_etl.py"})
    assert status == 200 and body["status"] == "succeeded"
    _, runs = _get(server + "/api/runs")
    assert any(r["pipeline"] == "simple-etl" for r in runs["runs"])
