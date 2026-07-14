# FlowForge - Integrations & Orchestration Demo

Complete end-to-end demonstration of FlowForge with Prefect, Dagster, and Terraform integration, including real-time observability and web UI.

## 🎯 Quick Start

### Prerequisites
- Python 3.9+
- Go 1.18+ (for web server)
- Git

### 1. Clone & Setup

```bash
cd FlowForge
python -m venv integrations/.venv
source integrations/.venv/bin/activate  # or .venv\Scripts\activate on Windows
pip install -r integrations/requirements.txt
```

### 2. Start Services

**Terminal 1 - Observability API:**
```bash
python -m uvicorn integrations.observability_api:app --host 127.0.0.1 --port 8000
```

**Terminal 2 - Web Server:**
```bash
go build -o web/server.exe ./web/server.go
./web/server.exe
```

### 3. Run Demos

**Terminal 3 - Prefect ETL:**
```bash
python integrations/prefect_flow.py
```

**Terminal 4 - Dagster ETL:**
```bash
python integrations/dagster_pipeline.py
```

### 4. View Live Demo
Open [http://localhost:8080](http://localhost:8080) in your browser and navigate to **Execution Logs** tab.

---

## 📁 File Structure

```
integrations/
├── prefect_flow.py           # Prefect ETL pipeline example
├── dagster_pipeline.py       # Dagster ETL job example
├── observability_api.py      # FastAPI observability service (SQLite)
├── scheduler_secrets.py      # Cron scheduling & Vault secrets mgmt
├── requirements.txt          # Python dependencies
├── docker-compose.yml        # Full Docker stack (optional)
├── Dockerfile                # Container image
├── README.md                 # This file
└── .venv/                    # Python virtual environment

web/
├── server.go                 # Go web server with UI
└── server.exe                # Compiled binary (Windows)
```

---

## 🔌 Components

### Observability API (`observability_api.py`)
FastAPI-based observability service that persists pipeline runs and task logs to SQLite.

**Endpoints:**
- `POST /runs` - Create/update a run record
- `POST /tasks` - Log task execution
- `GET /runs` - List all runs
- `GET /runs/{run_id}/tasks` - Get tasks for a run

**Database Schema:**
```sql
CREATE TABLE runs (
    id TEXT PRIMARY KEY,
    pipeline_id TEXT,
    status TEXT,
    started_at TEXT,
    finished_at TEXT,
    meta TEXT
);

CREATE TABLE task_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT,
    task_name TEXT,
    status TEXT,
    started_at TEXT,
    finished_at TEXT,
    logs TEXT
);
```

### Prefect Flow (`prefect_flow.py`)
Example ETL pipeline using Prefect with:
- **Extract** task: Simulates data extraction
- **Transform** task: Applies transformations with retry policy
- **Load** task: Writes to warehouse

**Features:**
- Automatic task logging to observability API
- Fallback mode if Prefect import fails
- Posts run/task events to observability API at http://localhost:8000

### Dagster Job (`dagster_pipeline.py`)
Example ETL job using Dagster with:
- **extract_op**: Data extraction operation
- **transform_op**: Data transformation
- **load_op**: Data loading

**Features:**
- Fallback execution mode (if Dagster import fails)
- Task-level logging to observability API
- Job composition

### Scheduler & Secrets (`scheduler_secrets.py`)
Production-ready secrets and scheduling module:

**Secrets Management:**
```python
from integrations.scheduler_secrets import SecretsManager

secrets = SecretsManager(vault_enabled=True, vault_addr="http://vault:8200")
db_password = secrets.get_secret("db-password")
```

**Pipeline Scheduling:**
```python
from integrations.scheduler_secrets import PipelineScheduler

scheduler = PipelineScheduler()
scheduler.add_schedule("etl-pipeline", "0 9 * * *")  # Daily at 9 AM
scheduler.trigger_pipeline("etl-pipeline")
```

### Web UI (`web/server.go`)
Interactive web interface with:
- **DAG Visualization** - Cytoscape.js-based pipeline diagram
- **Argo Workflows** - YAML generation for Kubernetes
- **Apache Airflow** - DAG Python code generation
- **Terraform** - Infrastructure-as-Code generation
- **Execution Logs** - Real-time run status and task logs (polls observability API every 5s)

---

## 🚀 Running the Full Stack

### Local Setup (No Docker)
1. Start observability API: `python -m uvicorn integrations.observability_api:app --host 127.0.0.1 --port 8000`
2. Build and run Go server: `go build -o web/server.exe ./web/server.go && ./web/server.exe`
3. Run Prefect demo: `python integrations/prefect_flow.py`
4. Run Dagster demo: `python integrations/dagster_pipeline.py`
5. Open browser to http://localhost:8080

### Docker Compose (Full Stack)
```bash
docker-compose -f integrations/docker-compose.yml up --build
```

Services:
- **FlowForge UI**: http://localhost:8080
- **Observability API**: http://localhost:8000
- **Prefect Orion**: http://localhost:4200 (optional)
- **Dagit**: http://localhost:3000 (optional)

---

## 📊 Demo Walkthrough

### 1. Prefect Demo
```bash
python integrations/prefect_flow.py
```
- Runs an ETL pipeline with extract → transform → load
- Posts run and task logs to observability API
- Demonstrates task dependencies and execution

### 2. Dagster Demo
```bash
python integrations/dagster_pipeline.py
```
- Executes a Dagster job
- Records execution to observability API
- Shows operator-based pipeline definition

### 3. View Results
Open http://localhost:8080, click **Execution Logs** tab to see:
- Pipeline status (running, completed, failed)
- Task breakdown with timestamps
- Logs from each task
- Metrics (CPU, memory, errors)
- Auto-updates every 5 seconds

---

## 🔐 Secrets Management

### Environment Variables (Development)
```bash
export DB_PASSWORD="mypassword"
export API_KEY="mykey123"
python integrations/scheduler_secrets.py
```

### HashiCorp Vault (Production)
```python
secrets = SecretsManager(
    vault_enabled=True,
    vault_addr="https://vault.company.com:8200",
    vault_token=os.getenv("VAULT_TOKEN")
)
password = secrets.get_secret("db-password")
```

---

## ⏰ Scheduling

### Add Pipeline Schedule
```python
from integrations.scheduler_secrets import PipelineScheduler

scheduler = PipelineScheduler()

# Daily at 9 AM
scheduler.add_schedule("etl-pipeline", "0 9 * * *")

# Every 6 hours
scheduler.add_schedule("analytics", "0 */6 * * *")

# List schedules
for pipeline_id, schedule in scheduler.list_schedules().items():
    print(f"{pipeline_id}: {schedule['cron']}")
```

### Manual Trigger
```python
scheduler.trigger_pipeline("etl-pipeline")
```

---

## 🛠️ Configuration

### Observability API
```python
# Set custom database path
DB_PATH = "/var/lib/flowforge/observability.db"
```

### Web Server
```go
// Change port in server.go
port := "9090"  // instead of "8080"
```

### Prefect Flow
```python
# Modify task retries
@task(retries=3, retry_delay_seconds=10)
def extract():
    ...
```

---

## 🐛 Troubleshooting

### Port Already in Use
```bash
# Find and kill process on port 8000
lsof -i :8000  # macOS/Linux
netstat -ano | findstr :8000  # Windows
taskkill /PID <PID> /F  # Windows
```

### Python Module Not Found
```bash
# Reinstall dependencies
pip install --upgrade pip
pip install -r requirements.txt
```

### Observability API Connection Refused
- Ensure API is running on port 8000
- Check firewall rules
- Verify localhost is accessible

---

## 📚 Documentation

- [Prefect Docs](https://docs.prefect.io)
- [Dagster Docs](https://docs.dagster.io)
- [Argo Workflows](https://argoproj.github.io/argo-workflows)
- [Apache Airflow](https://airflow.apache.org)
- [Terraform](https://www.terraform.io/docs)
- [HashiCorp Vault](https://www.vaultproject.io/docs)

---

## 📝 License

MIT License - See LICENSE file

---

**Happy Pipeline Building! 🚀**
