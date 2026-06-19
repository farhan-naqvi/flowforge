from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import sqlite3
from datetime import datetime
import os

DB_PATH = os.path.join(os.path.dirname(__file__), 'observability.db')

def init_db():
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    c.execute('''CREATE TABLE IF NOT EXISTS runs (
        id TEXT PRIMARY KEY,
        pipeline_id TEXT,
        status TEXT,
        started_at TEXT,
        finished_at TEXT,
        meta TEXT
    )''')
    c.execute('''CREATE TABLE IF NOT EXISTS task_logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        run_id TEXT,
        task_name TEXT,
        status TEXT,
        started_at TEXT,
        finished_at TEXT,
        logs TEXT
    )''')
    conn.commit()
    conn.close()

init_db()

app = FastAPI()

class RunPayload(BaseModel):
    id: str
    pipeline_id: str
    status: str
    started_at: str | None = None
    finished_at: str | None = None
    meta: str | None = None

class TaskPayload(BaseModel):
    run_id: str
    task_name: str
    status: str
    started_at: str | None = None
    finished_at: str | None = None
    logs: str | None = None

@app.post('/runs')
def create_run(r: RunPayload):
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    try:
        c.execute('INSERT OR REPLACE INTO runs (id,pipeline_id,status,started_at,finished_at,meta) VALUES (?,?,?,?,?,?)',
                  (r.id, r.pipeline_id, r.status, r.started_at, r.finished_at, r.meta))
        conn.commit()
    finally:
        conn.close()
    return {"ok": True}

@app.post('/tasks')
def create_task(t: TaskPayload):
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    try:
        c.execute('INSERT INTO task_logs (run_id,task_name,status,started_at,finished_at,logs) VALUES (?,?,?,?,?,?)',
                  (t.run_id, t.task_name, t.status, t.started_at, t.finished_at, t.logs))
        conn.commit()
    finally:
        conn.close()
    return {"ok": True}

@app.get('/runs')
def list_runs():
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    c.execute('SELECT id,pipeline_id,status,started_at,finished_at,meta FROM runs ORDER BY started_at DESC')
    rows = c.fetchall()
    conn.close()
    return [{"id": r[0], "pipeline_id": r[1], "status": r[2], "started_at": r[3], "finished_at": r[4], "meta": r[5]} for r in rows]

@app.get('/runs/{run_id}/tasks')
def get_tasks(run_id: str):
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    c.execute('SELECT task_name,status,started_at,finished_at,logs FROM task_logs WHERE run_id = ? ORDER BY id', (run_id,))
    rows = c.fetchall()
    conn.close()
    return [{"task_name": r[0], "status": r[1], "started_at": r[2], "finished_at": r[3], "logs": r[4]} for r in rows]
