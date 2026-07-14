"""SQLite-backed run/task event store.

Persists *real* execution events (no fabricated data). The schema is compatible
in spirit with the existing ``integrations/observability_api.py`` store so the
two can converge. Bounded field sizes guard against unbounded growth from
oversized logs.
"""

from __future__ import annotations

import os
import sqlite3
import threading
import time
from contextlib import contextmanager
from typing import Any, Dict, List, Optional

MAX_LOG_CHARS = 64 * 1024  # cap persisted per-task log size

DEFAULT_DB = os.environ.get("FLOWFORGE_DB", os.path.expanduser("~/.flowforge/runs.db"))

_SCHEMA = """
CREATE TABLE IF NOT EXISTS runs (
    id TEXT PRIMARY KEY,
    pipeline TEXT NOT NULL,
    status TEXT NOT NULL,
    mode TEXT NOT NULL,
    started_at REAL,
    finished_at REAL
);
CREATE TABLE IF NOT EXISTS task_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    task TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at REAL,
    finished_at REAL,
    duration_ms INTEGER,
    logs TEXT,
    error TEXT,
    FOREIGN KEY (run_id) REFERENCES runs(id)
);
"""


class Store:
    def __init__(self, path: str = DEFAULT_DB) -> None:
        self.path = path
        if path != ":memory:":
            os.makedirs(os.path.dirname(os.path.abspath(path)), exist_ok=True)
        # check_same_thread=False + a lock: the local server serves requests on
        # worker threads, so the single connection is shared under the lock.
        self._conn = sqlite3.connect(path, check_same_thread=False)
        self._conn.row_factory = sqlite3.Row
        self._lock = threading.Lock()
        self._conn.executescript(_SCHEMA)
        self._conn.commit()

    @contextmanager
    def _cursor(self):
        with self._lock:
            cur = self._conn.cursor()
            try:
                yield cur
                self._conn.commit()
            finally:
                cur.close()

    def close(self) -> None:
        self._conn.close()

    # -- writes ----------------------------------------------------------

    def start_run(self, run_id: str, pipeline: str, mode: str) -> None:
        with self._cursor() as cur:
            cur.execute(
                "INSERT OR REPLACE INTO runs (id, pipeline, status, mode, started_at) VALUES (?,?,?,?,?)",
                (run_id, pipeline, "running", mode, time.time()),
            )

    def finish_run(self, run_id: str, status: str) -> None:
        with self._cursor() as cur:
            cur.execute(
                "UPDATE runs SET status=?, finished_at=? WHERE id=?",
                (status, time.time(), run_id),
            )

    def record_task(
        self,
        run_id: str,
        task: str,
        status: str,
        started_at: float,
        finished_at: float,
        logs: str = "",
        error: str = "",
    ) -> None:
        logs = (logs or "")[:MAX_LOG_CHARS]
        duration_ms = int((finished_at - started_at) * 1000)
        with self._cursor() as cur:
            cur.execute(
                "INSERT INTO task_events (run_id, task, status, started_at, finished_at, duration_ms, logs, error)"
                " VALUES (?,?,?,?,?,?,?,?)",
                (run_id, task, status, started_at, finished_at, duration_ms, logs, error[:MAX_LOG_CHARS]),
            )

    # -- reads -----------------------------------------------------------

    def list_runs(self, limit: int = 50) -> List[Dict[str, Any]]:
        with self._cursor() as cur:
            cur.execute(
                "SELECT id, pipeline, status, mode, started_at, finished_at FROM runs"
                " ORDER BY started_at DESC LIMIT ?",
                (limit,),
            )
            return [dict(r) for r in cur.fetchall()]

    def get_run(self, run_id: str) -> Optional[Dict[str, Any]]:
        with self._cursor() as cur:
            cur.execute(
                "SELECT id, pipeline, status, mode, started_at, finished_at FROM runs WHERE id=?",
                (run_id,),
            )
            row = cur.fetchone()
            if not row:
                return None
            run = dict(row)
            cur.execute(
                "SELECT task, status, started_at, finished_at, duration_ms, logs, error"
                " FROM task_events WHERE run_id=? ORDER BY id",
                (run_id,),
            )
            run["tasks"] = [dict(r) for r in cur.fetchall()]
            return run
