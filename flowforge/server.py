"""Local workbench server: JSON API + static UI, Python stdlib only.

Security posture (see docs/security/THREAT_MODEL.md):
- binds 127.0.0.1 by default (explicit --host required to expose);
- rejects requests whose Host header is not loopback (DNS-rebinding guard);
- serves and executes only files inside the configured workspace (traversal guard);
- caps request bodies.

The /api/run endpoint executes the user's own pipeline code in-process. This is
a local developer tool, not a sandbox.
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any, Dict, Optional
from urllib.parse import parse_qs, urlparse

from .capability import analyze_all
from .compilers import TARGETS
from .compilers import compile as compile_target
from .parser import ParseError, parse_file
from .store import Store
from .validation import validate

WEBUI_DIR = os.path.join(os.path.dirname(__file__), "webui")
MAX_BODY = 256 * 1024
_ALLOWED_HOST_SUFFIXES = ("localhost", "127.0.0.1", "::1")

_CONTENT_TYPES = {".html": "text/html; charset=utf-8", ".js": "text/javascript; charset=utf-8",
                  ".css": "text/css; charset=utf-8"}


class _Context:
    def __init__(self, workspace: str, store: Store) -> None:
        self.workspace = os.path.abspath(workspace)
        self.store = store

    def resolve(self, rel: str) -> Optional[str]:
        """Resolve a workspace-relative path, rejecting traversal."""
        if not rel:
            return None
        candidate = os.path.abspath(os.path.join(self.workspace, rel))
        if candidate != self.workspace and not candidate.startswith(self.workspace + os.sep):
            return None
        return candidate if os.path.isfile(candidate) else None

    def list_pipelines(self) -> list:
        out = []
        for root, _dirs, files in os.walk(self.workspace):
            if any(part in (".git", "node_modules", "__pycache__", ".venv") for part in root.split(os.sep)):
                continue
            for fn in sorted(files):
                if not fn.endswith(".py"):
                    continue
                path = os.path.join(root, fn)
                try:
                    spec = parse_file(path)
                except (ParseError, OSError, SyntaxError):
                    continue
                rel = os.path.relpath(path, self.workspace)
                out.append({
                    "file": rel,
                    "name": spec.metadata.name,
                    "tasks": len(spec.tasks),
                    "edges": len(spec.edges),
                })
        return out


class Handler(BaseHTTPRequestHandler):
    ctx: _Context = None  # set by serve()
    server_version = "FlowForge"

    # -- helpers ---------------------------------------------------------

    def _host_ok(self) -> bool:
        host = (self.headers.get("Host") or "").split(":")[0]
        return host == "" or host in _ALLOWED_HOST_SUFFIXES

    def _json(self, obj: Any, code: int = 200) -> None:
        body = json.dumps(obj).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("X-Content-Type-Options", "nosniff")
        self.send_header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'")
        self.end_headers()
        self.wfile.write(body)

    def _error(self, code: int, message: str) -> None:
        self._json({"error": message}, code)

    def _static(self, name: str) -> None:
        safe = os.path.basename(name) or "index.html"
        path = os.path.join(WEBUI_DIR, safe)
        if not os.path.isfile(path):
            self._error(404, "not found")
            return
        with open(path, "rb") as fh:
            body = fh.read()
        ext = os.path.splitext(safe)[1]
        self.send_response(200)
        self.send_header("Content-Type", _CONTENT_TYPES.get(ext, "application/octet-stream"))
        self.send_header("Content-Length", str(len(body)))
        self.send_header("X-Content-Type-Options", "nosniff")
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, fmt, *args):  # quieter logging
        return

    # -- routing ---------------------------------------------------------

    def do_GET(self):  # noqa: N802
        if not self._host_ok():
            self._error(403, "host not allowed")
            return
        parsed = urlparse(self.path)
        path = parsed.path
        q = parse_qs(parsed.query)
        try:
            if path in ("/", "/index.html"):
                self._static("index.html")
            elif path in ("/app.js", "/styles.css"):
                self._static(path.lstrip("/"))
            elif path == "/api/health":
                self._json({"status": "ok", "workspace": self.ctx.workspace, "db": self.ctx.store.path})
            elif path == "/api/pipelines":
                self._json({"pipelines": self.ctx.list_pipelines()})
            elif path == "/api/pipeline":
                self._api_pipeline(q.get("file", [""])[0])
            elif path == "/api/compile":
                self._api_compile(q.get("file", [""])[0], q.get("target", [""])[0])
            elif path == "/api/runs":
                self._json({"runs": self.ctx.store.list_runs()})
            elif path == "/api/run":
                run = self.ctx.store.get_run(q.get("id", [""])[0])
                self._json(run) if run else self._error(404, "run not found")
            else:
                self._error(404, "not found")
        except Exception as exc:  # noqa: BLE001
            self._error(500, str(exc))

    def do_POST(self):  # noqa: N802
        if not self._host_ok():
            self._error(403, "host not allowed")
            return
        length = int(self.headers.get("Content-Length") or 0)
        if length > MAX_BODY:
            self._error(413, "payload too large")
            return
        raw = self.rfile.read(length) if length else b"{}"
        try:
            payload = json.loads(raw or b"{}")
        except json.JSONDecodeError:
            self._error(400, "invalid JSON")
            return
        if urlparse(self.path).path == "/api/run":
            self._api_run(payload)
        else:
            self._error(404, "not found")

    # -- endpoints -------------------------------------------------------

    def _api_pipeline(self, file: str) -> None:
        path = self.ctx.resolve(file)
        if not path:
            self._error(404, "pipeline file not found in workspace")
            return
        try:
            spec = parse_file(path)
        except ParseError as exc:
            self._error(400, f"parse error: {exc}")
            return
        self._json({
            "file": file,
            "ir": spec.to_dict(),
            "validation": validate(spec).to_dict(),
            "capability": {t: r.to_dict() for t, r in analyze_all(spec).items()},
        })

    def _api_compile(self, file: str, target: str) -> None:
        path = self.ctx.resolve(file)
        if not path:
            self._error(404, "pipeline file not found in workspace")
            return
        if target not in TARGETS:
            self._error(400, f"unknown target {target!r}")
            return
        spec = parse_file(path)
        result = validate(spec)
        if not result.ok:
            self._json({"ok": False, "validation": result.to_dict()})
            return
        self._json({"ok": True, "target": target, "artifact": compile_target(spec, target)})

    def _api_run(self, payload: Dict[str, Any]) -> None:
        from .loader import LoadError, load_pipeline
        from .runner import run_pipeline

        file = payload.get("file", "")
        path = self.ctx.resolve(file)
        if not path:
            self._error(404, "pipeline file not found in workspace")
            return
        try:
            pipeline = load_pipeline(path, name=payload.get("name"))
        except LoadError as exc:
            self._error(400, str(exc))
            return
        result = run_pipeline(pipeline, store=self.ctx.store)
        self._json(result.to_dict())


def serve(host: str = "127.0.0.1", port: int = 8080, db: Optional[str] = None, workspace: str = ".") -> None:
    store = Store(db) if db else Store()
    Handler.ctx = _Context(workspace, store)
    httpd = ThreadingHTTPServer((host, port), Handler)
    exposed = host not in ("127.0.0.1", "localhost", "::1")
    print(f"FlowForge workbench on http://{host}:{port}  (workspace: {Handler.ctx.workspace})")
    if exposed:
        print("WARNING: bound to a non-loopback address; the local-run endpoint executes code. "
              "Only do this on a trusted network.")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\nshutting down")
    finally:
        httpd.server_close()
        store.close()
