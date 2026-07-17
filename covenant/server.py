"""Local Covenant workbench server: JSON API + static UI (stdlib only).

Same security posture as the rest of Covenant's local tooling: binds 127.0.0.1
by default, rejects non-loopback Host headers, caps request bodies. The server
reads the committed ``covenant-contracts`` tree and can write a plan file (a
PR-ready change), but performs no production actions.
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Any, Optional
from urllib.parse import parse_qs, urlparse

from covenant_transforms import all_ids as primitive_ids
from covenant_transforms.schema import DTYPES

from . import authoring, gitops
from .authoring import AuthoringError
from .compile import compile_argo
from .model import Intent
from .odcs import load_contract
from .planner import PlanError, plan_from_intent
from .verify import verify_plan

WEBUI_DIR = os.path.join(os.path.dirname(__file__), "webui")
MAX_BODY = 256 * 1024
_LOOPBACK = ("localhost", "127.0.0.1", "::1", "")
_CT = {".html": "text/html; charset=utf-8", ".js": "text/javascript; charset=utf-8",
       ".css": "text/css; charset=utf-8"}


class _Ctx:
    def __init__(self, base_dir: str) -> None:
        self.base_dir = os.path.abspath(base_dir)


def _product_payload(slug: str, base_dir: str) -> dict:
    products = {p.slug: p for p in gitops.discover(base_dir)}
    if slug not in products:
        raise KeyError(slug)
    intent = Intent.load(os.path.join(base_dir, products[slug].intent_path))
    source = load_contract(os.path.join(base_dir, intent.source_contract))
    target = load_contract(os.path.join(base_dir, intent.target_contract))
    payload = {
        "slug": slug,
        "source": {"id": source.id, "schema": source.schema.to_dicts(), "primary_key": source.primary_key},
        "target": {"id": target.id, "schema": target.schema.to_dicts(), "primary_key": target.primary_key},
        "intent": [{"primitive": s.primitive, "params": s.params} for s in intent.steps],
        "plan": None,
        "plan_error": None,
    }
    # Planning can fail while a pipeline is being authored (a bad step, an
    # unknown column). The product must still load so the user can fix it.
    try:
        payload["plan"] = plan_from_intent(intent, base_dir=base_dir).to_dict()
    except PlanError as exc:
        payload["plan_error"] = str(exc)
    return payload


class Handler(BaseHTTPRequestHandler):
    ctx: _Ctx = None
    server_version = "Covenant"

    def _host_ok(self) -> bool:
        return (self.headers.get("Host") or "").split(":")[0] in _LOOPBACK

    def _json(self, obj: Any, code: int = 200) -> None:
        body = json.dumps(obj).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("X-Content-Type-Options", "nosniff")
        self.send_header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'")
        self.end_headers()
        self.wfile.write(body)

    def _err(self, code: int, msg: str) -> None:
        self._json({"error": msg}, code)

    def _static(self, name: str) -> None:
        safe = os.path.basename(name) or "index.html"
        path = os.path.join(WEBUI_DIR, safe)
        if not os.path.isfile(path):
            return self._err(404, "not found")
        with open(path, "rb") as fh:
            body = fh.read()
        self.send_response(200)
        self.send_header("Content-Type", _CT.get(os.path.splitext(safe)[1], "application/octet-stream"))
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *a):
        return

    def do_GET(self):  # noqa: N802
        if not self._host_ok():
            return self._err(403, "host not allowed")
        u = urlparse(self.path)
        q = parse_qs(u.query)
        base = self.ctx.base_dir
        try:
            if u.path in ("/", "/index.html"):
                self._static("index.html")
            elif u.path in ("/app.js", "/styles.css"):
                self._static(u.path.lstrip("/"))
            elif u.path == "/api/health":
                self._json({"status": "ok", "base_dir": base})
            elif u.path == "/api/meta":
                self._json({"dtypes": list(DTYPES), "primitives": primitive_ids()})
            elif u.path == "/api/products":
                self._json({"products": [p.__dict__ for p in gitops.discover(base)]})
            elif u.path == "/api/product":
                self._json(_product_payload(q.get("slug", [""])[0], base))
            elif u.path == "/api/compile":
                self._api_compile(q.get("slug", [""])[0], base)
            else:
                self._err(404, "not found")
        except KeyError:
            self._err(404, "data product not found")
        except (PlanError, FileNotFoundError) as exc:
            self._err(400, str(exc))
        except Exception as exc:  # noqa: BLE001
            self._err(500, str(exc))

    def do_POST(self):  # noqa: N802
        if not self._host_ok():
            return self._err(403, "host not allowed")
        n = int(self.headers.get("Content-Length") or 0)
        if n > MAX_BODY:
            return self._err(413, "payload too large")
        try:
            payload = json.loads(self.rfile.read(n) or b"{}")
        except json.JSONDecodeError:
            return self._err(400, "invalid JSON")
        u = urlparse(self.path)
        base = self.ctx.base_dir
        try:
            if u.path == "/api/verify":
                self._api_verify(payload.get("slug", ""), base)
            elif u.path == "/api/plan/save":
                self._api_save(payload.get("slug", ""), base)
            elif u.path == "/api/product/create":
                self._api_create(payload.get("slug", ""), base)
            elif u.path == "/api/contract/save":
                self._api_contract_save(payload, base)
            elif u.path == "/api/intent/save":
                self._api_intent_save(payload, base)
            else:
                self._err(404, "not found")
        except AuthoringError as exc:
            self._err(400, str(exc))  # visible validation error
        except KeyError:
            self._err(404, "data product not found")
        except Exception as exc:  # noqa: BLE001
            self._err(500, str(exc))

    # -- endpoints -------------------------------------------------------

    def _api_verify(self, slug: str, base: str) -> None:
        products = {p.slug: p for p in gitops.discover(base)}
        intent = Intent.load(os.path.join(base, products[slug].intent_path))
        plan = plan_from_intent(intent, base_dir=base)
        target = load_contract(os.path.join(base, plan.target_contract))
        verdict = verify_plan(plan, target, base_dir=base)
        self._json({"static": plan.conformance.to_dict(), "dynamic": verdict.to_dict()})

    def _api_compile(self, slug: str, base: str) -> None:
        products = {p.slug: p for p in gitops.discover(base)}
        intent = Intent.load(os.path.join(base, products[slug].intent_path))
        plan = plan_from_intent(intent, base_dir=base)
        if not plan.conformance.ok:
            return self._json({"ok": False, "conformance": plan.conformance.to_dict()})
        art = compile_argo(plan, plan_path=gitops.plan_path(slug).replace(base + os.sep, ""))
        self._json({"ok": True, "artifact": art})

    def _api_save(self, slug: str, base: str) -> None:
        products = {p.slug: p for p in gitops.discover(base)}
        intent = Intent.load(os.path.join(base, products[slug].intent_path))
        plan = plan_from_intent(intent, base_dir=base)
        if not plan.conformance.ok:
            return self._json({"ok": False, "reason": "plan does not conform; not saved"})
        rel = gitops.write_plan(slug, plan.to_yaml(), base)
        self._json({"ok": True, "path": rel, "note": "commit on a branch and open a PR (CODEOWNERS-gated)"})

    # -- authoring (write side) -----------------------------------------

    def _api_create(self, slug: str, base: str) -> None:
        result = authoring.create_product(base, slug)  # AuthoringError if exists/invalid
        self._json({"ok": True, **result})

    def _api_contract_save(self, payload: dict, base: str) -> None:
        slug = payload.get("slug", "")
        kind = payload.get("kind", "")
        fields = payload.get("fields", [])
        if not authoring.product_exists(base, slug):
            return self._err(404, f"data product {slug!r} does not exist; create it first")
        rel = authoring.save_contract(base, slug, kind, fields)  # AuthoringError -> 400
        # Return the freshly re-planned product so the UI updates conformance now.
        self._json({"ok": True, "path": rel, "product": _product_payload(slug, base)})

    def _api_intent_save(self, payload: dict, base: str) -> None:
        slug = payload.get("slug", "")
        steps = payload.get("steps", [])
        if not authoring.product_exists(base, slug):
            return self._err(404, f"data product {slug!r} does not exist; create it first")
        rel = authoring.save_intent(base, slug, steps)  # AuthoringError -> 400
        self._json({"ok": True, "path": rel, "product": _product_payload(slug, base)})


def serve(host: str = "127.0.0.1", port: int = 8090, base_dir: str = ".") -> None:
    Handler.ctx = _Ctx(base_dir)
    # Single-threaded on purpose: this is a local single-user tool, and the
    # embedded DuckDB used by verify is not safe under concurrent request
    # threads. Requests serialize (each is fast); no cross-thread native crashes.
    httpd = HTTPServer((host, port), Handler)
    print(f"Covenant workbench on http://{host}:{port}  (contracts: {Handler.ctx.base_dir})")
    if host not in ("127.0.0.1", "localhost", "::1"):
        print("WARNING: bound to a non-loopback address.")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\nshutting down")
    finally:
        httpd.server_close()
