"""Server-level tests for the authoring endpoints, against a throwaway
base_dir so the repo's committed products are never touched."""

import json
import threading
import urllib.error
import urllib.request
from http.server import ThreadingHTTPServer

import pytest

from covenant.server import Handler, _Ctx


@pytest.fixture()
def server(tmp_path):
    Handler.ctx = _Ctx(str(tmp_path))
    httpd = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    port = httpd.server_address[1]
    t = threading.Thread(target=httpd.serve_forever, daemon=True)
    t.start()
    yield f"http://127.0.0.1:{port}"
    httpd.shutdown()


def _get(url):
    with urllib.request.urlopen(url) as r:
        return r.status, json.loads(r.read())


def _post(url, payload):
    req = urllib.request.Request(url, data=json.dumps(payload).encode(),
                                 headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req) as r:
        return r.status, json.loads(r.read())


def test_meta_lists_types_and_primitives(server):
    _, body = _get(server + "/api/meta")
    assert "string" in body["dtypes"] and "filter" in body["primitives"]


def test_create_then_appears_in_products(server):
    status, body = _post(server + "/api/product/create", {"slug": "sales/orders"})
    assert status == 200 and body["ok"]
    _, prods = _get(server + "/api/products")
    assert "sales/orders" in [p["slug"] for p in prods["products"]]


def test_create_duplicate_is_visible_error(server):
    _post(server + "/api/product/create", {"slug": "a/b"})
    with pytest.raises(urllib.error.HTTPError) as exc:
        _post(server + "/api/product/create", {"slug": "a/b"})
    assert exc.value.code == 400


def test_save_contract_returns_replanned_product(server):
    _post(server + "/api/product/create", {"slug": "sales/orders"})
    status, body = _post(server + "/api/contract/save", {
        "slug": "sales/orders", "kind": "target",
        "fields": [{"name": "d", "type": "date", "nullable": False, "primary_key": True}],
    })
    assert status == 200 and body["ok"]
    assert "plan" in body["product"]  # re-planned payload returned for immediate UI update


def test_save_invalid_contract_is_visible_error(server):
    _post(server + "/api/product/create", {"slug": "a/b"})
    with pytest.raises(urllib.error.HTTPError) as exc:
        _post(server + "/api/contract/save", {"slug": "a/b", "kind": "source",
                                              "fields": [{"name": "x", "type": "bogus"}]})
    assert exc.value.code == 400


def test_contract_save_requires_existing_product(server):
    with pytest.raises(urllib.error.HTTPError) as exc:
        _post(server + "/api/contract/save", {"slug": "ghost/x", "kind": "source",
                                              "fields": [{"name": "a", "type": "string"}]})
    assert exc.value.code == 404
