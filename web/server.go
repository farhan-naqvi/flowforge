package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Create a sample pipeline
	pipeline := createSamplePipeline()

	// Serve static files (HTML, CSS, JS)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(getHTMLContent()))
	})

	// API endpoint for pipeline data
	http.HandleFunc("/api/pipeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pipeline)
	})

	// API endpoint for Argo Workflow output
	http.HandleFunc("/api/argo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(generateArgoWorkflow(pipeline))
	})

	// API endpoint for Airflow DAG output
	http.HandleFunc("/api/airflow", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(generateAirflowDAG(pipeline))
	})

	// API endpoint for Terraform output
	http.HandleFunc("/api/terraform", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(generateTerraform(pipeline))
	})

	// API endpoint for execution logs
	// API endpoint for execution logs (proxy to observability service if available)
	http.HandleFunc("/api/execution", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// try observability service
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://localhost:8000/runs")
		if err != nil || resp.StatusCode != 200 {
			json.NewEncoder(w).Encode(generateExecutionLogs())
			return
		}
		defer resp.Body.Close()
		var runs []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil || len(runs) == 0 {
			json.NewEncoder(w).Encode(generateExecutionLogs())
			return
		}
		// pick most recent run
		run := runs[0]
		runID, _ := run["id"].(string)
		// fetch tasks
		tresp, err := client.Get("http://localhost:8000/runs/" + runID + "/tasks")
		if err != nil || tresp.StatusCode != 200 {
			json.NewEncoder(w).Encode(generateExecutionLogs())
			return
		}
		defer tresp.Body.Close()
		var tasks []map[string]interface{}
		if err := json.NewDecoder(tresp.Body).Decode(&tasks); err != nil {
			json.NewEncoder(w).Encode(generateExecutionLogs())
			return
		}
		// map to expected structure
		out := map[string]interface{}{
			"pipeline_id": run["pipeline_id"],
			"status":      run["status"],
			"started_at":  run["started_at"],
			"tasks":       []interface{}{},
			"metrics":     map[string]interface{}{},
		}
		for _, t := range tasks {
			outTasks := out["tasks"].([]interface{})
			outTasks = append(outTasks, map[string]interface{}{
				"name":     t["task_name"],
				"status":   t["status"],
				"started":  t["started_at"],
				"completed": t["finished_at"],
				"duration": "-",
				"logs":     t["logs"],
			})
			out["tasks"] = outTasks
		}
		json.NewEncoder(w).Encode(out)
	})

	port := "8080"
	fmt.Printf("🚀 FlowForge UI Server running at http://localhost:%s\n", port)
	fmt.Printf("📊 DAG Visualization: http://localhost:%s\n", port)
	fmt.Printf("✋ Press Ctrl+C to stop\n\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func createSamplePipeline() map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "flowforge.io/v1",
		"kind":       "Pipeline",
		"metadata": map[string]interface{}{
			"name":        "etl-pipeline",
			"version":     "1.0.0",
			"namespace":   "default",
			"description": "Sample ETL pipeline demonstrating FlowForge capabilities",
			"owner":       "data-team",
			"tags": map[string]string{
				"team":        "analytics",
				"environment": "staging",
				"type":        "etl",
			},
		},
		"tasks": map[string]interface{}{
			"extract": map[string]interface{}{
				"type":        "Source",
				"description": "Extract data from source database",
				"handler": map[string]interface{}{
					"type":   "python",
					"source": "s3://bucket/extract.py",
				},
				"outputs": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
					},
				},
			},
			"transform": map[string]interface{}{
				"type":        "Transform",
				"description": "Transform and clean data",
				"handler": map[string]interface{}{
					"type":   "python",
					"source": "s3://bucket/transform.py",
				},
				"inputs": map[string]interface{}{
					"input": map[string]interface{}{
						"type": "object",
					},
				},
				"outputs": map[string]interface{}{
					"output": map[string]interface{}{
						"type": "object",
					},
				},
				"timeout": "1h",
				"retry": map[string]interface{}{
					"maxAttempts": 3,
				},
			},
			"load": map[string]interface{}{
				"type":        "Sink",
				"description": "Load transformed data to data warehouse",
				"handler": map[string]interface{}{
					"type":   "python",
					"source": "s3://bucket/load.py",
				},
				"inputs": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
					},
				},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"from": map[string]interface{}{
					"task": "extract",
					"port": "data",
				},
				"to": map[string]interface{}{
					"task": "transform",
					"port": "input",
				},
			},
			map[string]interface{}{
				"from": map[string]interface{}{
					"task": "transform",
					"port": "output",
				},
				"to": map[string]interface{}{
					"task": "load",
					"port": "data",
				},
			},
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

func generateArgoWorkflow(pipeline map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Workflow",
		"metadata": map[string]interface{}{
			"generateName": "etl-pipeline-",
			"namespace":    "argo",
		},
		"spec": map[string]interface{}{
			"entrypoint": "dag",
			"templates": []interface{}{
				map[string]interface{}{
					"name": "dag",
					"dag": map[string]interface{}{
						"tasks": []interface{}{
							map[string]interface{}{
								"name":     "extract",
								"template": "extract-task",
							},
							map[string]interface{}{
								"name":     "transform",
								"template": "transform-task",
								"depends":  "extract",
							},
							map[string]interface{}{
								"name":     "load",
								"template": "load-task",
								"depends":  "transform",
							},
						},
					},
				},
				map[string]interface{}{
					"name": "extract-task",
					"container": map[string]interface{}{
						"image":   "python:3.9",
						"command": []string{"python", "s3://bucket/extract.py"},
						"env": []interface{}{
							map[string]interface{}{"name": "AWS_REGION", "value": "us-east-1"},
						},
						"resources": map[string]interface{}{
							"requests": map[string]interface{}{"cpu": "500m", "memory": "512Mi"},
						},
					},
				},
				map[string]interface{}{
					"name": "transform-task",
					"container": map[string]interface{}{
						"image":   "python:3.9",
						"command": []string{"python", "s3://bucket/transform.py"},
						"resources": map[string]interface{}{
							"requests": map[string]interface{}{"cpu": "1", "memory": "2Gi"},
						},
					},
					"retryStrategy": map[string]interface{}{
						"limit": 3,
						"backoff": map[string]interface{}{
							"duration":    "5s",
							"factor":      2,
							"maxDuration": "5m",
						},
					},
				},
				map[string]interface{}{
					"name": "load-task",
					"container": map[string]interface{}{
						"image":   "python:3.9",
						"command": []string{"python", "s3://bucket/load.py"},
						"resources": map[string]interface{}{
							"requests": map[string]interface{}{"cpu": "2", "memory": "4Gi"},
						},
					},
				},
			},
			"activeDeadlineSeconds": 3600,
		},
		"status": "Submitted",
	}
}

func generateAirflowDAG(pipeline map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"dag_id":            "etl-pipeline",
		"description":       "Sample ETL pipeline demonstrating FlowForge capabilities",
		"schedule_interval": "@daily",
		"default_view":      "graph",
		"tasks": []interface{}{
			map[string]interface{}{
				"task_id":         "extract",
				"operator":        "PythonOperator",
				"python_callable": "extract_data",
				"op_args":         []interface{}{},
			},
			map[string]interface{}{
				"task_id":         "transform",
				"operator":        "PythonOperator",
				"python_callable": "transform_data",
				"upstream":        []string{"extract"},
				"retries":         3,
				"retry_delay":     "5m",
			},
			map[string]interface{}{
				"task_id":         "load",
				"operator":        "PythonOperator",
				"python_callable": "load_data",
				"upstream":        []string{"transform"},
			},
		},
		"code_sample": `from airflow import DAG
from airflow.operators.python import PythonOperator
from datetime import datetime, timedelta

default_args = {
    'owner': 'data-team',
    'retries': 3,
    'retry_delay': timedelta(minutes=5),
}

dag = DAG('etl-pipeline', default_args=default_args, schedule_interval='@daily')

def extract_data():
    print("Extracting data from source...")

def transform_data():
    print("Transforming and cleaning data...")

def load_data():
    print("Loading data to warehouse...")

t1 = PythonOperator(task_id='extract', python_callable=extract_data, dag=dag)
t2 = PythonOperator(task_id='transform', python_callable=transform_data, dag=dag)
t3 = PythonOperator(task_id='load', python_callable=load_data, dag=dag)

t1 >> t2 >> t3`,
	}
}

func generateTerraform(pipeline map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"resource": "aws_ecs_task_definition",
		"name":     "etl-pipeline",
		"hcl_code": `resource "aws_ecs_task_definition" "etl_pipeline" {
  family                   = "etl-pipeline"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "2048"
  memory                   = "4096"

  container_definitions = jsonencode([
    {
      name      = "extract"
      image     = "python:3.9"
      cpu       = 512
      memory    = 512
      essential = true
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = "/ecs/etl-pipeline"
          "awslogs-region"        = "us-east-1"
          "awslogs-stream-prefix" = "ecs"
        }
      }
    },
    {
      name      = "transform"
      image     = "python:3.9"
      cpu       = 1024
      memory    = 2048
      essential = true
    },
    {
      name      = "load"
      image     = "python:3.9"
      cpu       = 512
      memory    = 1024
      essential = true
    }
  ])
}

resource "helm_release" "etl_pipeline" {
  name       = "etl-pipeline"
  repository = "argo"
  chart      = "argo-workflows"
  namespace  = "argo"

  values = [
    file("${path.module}/values.yaml")
  ]
}`,
	}
}

func generateExecutionLogs() map[string]interface{} {
	return map[string]interface{}{
		"pipeline_id": "etl-pipeline-001",
		"status":      "running",
		"started_at":  time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		"tasks": []interface{}{
			map[string]interface{}{
				"name":      "extract",
				"status":    "completed",
				"started":   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
				"completed": time.Now().Add(-4 * time.Minute).Format(time.RFC3339),
				"duration":  "1m 23s",
				"logs":      "Extracted 15,234 records from database\nData format: CSV\nSize: 234 MB",
			},
			map[string]interface{}{
				"name":     "transform",
				"status":   "running",
				"started":  time.Now().Add(-4 * time.Minute).Format(time.RFC3339),
				"progress": "75%",
				"duration": "2m 15s (elapsed)",
				"logs":     "Processing records...\nApplied 8 transformations\nRemoved 234 duplicates\n14,890 records remaining",
			},
			map[string]interface{}{
				"name":      "load",
				"status":    "pending",
				"scheduled": time.Now().Add(2 * time.Minute).Format(time.RFC3339),
				"logs":      "Waiting for transform to complete...",
			},
		},
		"metrics": map[string]interface{}{
			"cpu_usage":      "45%",
			"memory_usage":   "2.1 GB",
			"data_processed": "14.8 MB/s",
			"errors":         0,
		},
	}
}

func getHTMLContent() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FlowForge - Complete Demo</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/cytoscape/3.28.1/cytoscape.min.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0f172a; color: #e2e8f0; height: 100vh; display: flex; flex-direction: column; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 20px 30px; box-shadow: 0 4px 6px rgba(0,0,0,0.3); z-index: 100; }
        .header h1 { font-size: 28px; font-weight: 600; display: flex; align-items: center; gap: 12px; }
        .header p { font-size: 14px; margin-top: 6px; opacity: 0.9; }
        .tabs { display: flex; gap: 0; border-bottom: 2px solid #334155; background: #1e293b; padding: 0 20px; }
        .tab-btn { background: transparent; border: none; color: #94a3b8; cursor: pointer; padding: 15px 20px; font-size: 14px; font-weight: 500; border-bottom: 3px solid transparent; transition: all 0.3s ease; }
        .tab-btn:hover { color: #60a5fa; }
        .tab-btn.active { color: #667eea; border-bottom-color: #667eea; }
        .container { display: flex; flex: 1; overflow: hidden; }
        .content { display: flex; flex: 1; gap: 20px; padding: 20px; overflow: hidden; }
        .content.hidden { display: none; }
        #cy { flex: 1; background: #1e293b; border-radius: 8px; border: 1px solid #334155; }
        .sidebar { width: 300px; background: #1e293b; border-radius: 8px; border: 1px solid #334155; overflow-y: auto; padding: 15px; }
        .code-viewer { flex: 1; background: #1e293b; border-radius: 8px; border: 1px solid #334155; overflow: auto; padding: 15px; font-family: 'Monaco', 'Courier New', monospace; }
        .code-block { background: #0f172a; border: 1px solid #334155; border-radius: 6px; padding: 15px; margin-bottom: 15px; overflow-x: auto; font-size: 12px; line-height: 1.5; }
        pre { margin: 0; white-space: pre-wrap; word-wrap: break-word; }
        .execution-view { display: flex; flex-direction: column; gap: 15px; }
        .task-status { background: #0f172a; border: 1px solid #334155; border-radius: 6px; padding: 15px; }
        .task-status.completed { border-left: 4px solid #10b981; }
        .task-status.running { border-left: 4px solid #667eea; }
        .task-status.pending { border-left: 4px solid #f59e0b; }
        .task-status-header { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
        .status-badge { padding: 3px 8px; border-radius: 3px; font-size: 11px; font-weight: 600; text-transform: uppercase; }
        .badge-completed { background: #10b981; color: #000; }
        .badge-running { background: #667eea; color: #fff; }
        .badge-pending { background: #f59e0b; color: #000; }
        .progress-bar { background: #334155; height: 4px; border-radius: 2px; margin: 10px 0; overflow: hidden; }
        .progress-fill { background: #667eea; height: 100%; width: 75%; border-radius: 2px; }
        .task-logs { background: #000; border: 1px solid #334155; border-radius: 4px; padding: 10px; font-size: 11px; color: #10b981; font-family: 'Monaco', monospace; line-height: 1.4; max-height: 200px; overflow-y: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 FlowForge - Complete End-to-End Demo</h1>
        <p>Interactive pipeline with Argo Workflows, Apache Airflow, Terraform, and live execution</p>
    </div>
    <div class="tabs">
        <button class="tab-btn active" onclick="switchTab('dag')">📊 DAG Visualization</button>
        <button class="tab-btn" onclick="switchTab('argo')">⚙️ Argo Workflows</button>
        <button class="tab-btn" onclick="switchTab('airflow')">🔄 Apache Airflow</button>
        <button class="tab-btn" onclick="switchTab('terraform')">🏗️ Terraform</button>
        <button class="tab-btn" onclick="switchTab('execution')">▶️ Execution Logs</button>
    </div>
    <div class="container">
        <div id="dag" class="content">
            <div id="cy"></div>
            <div class="sidebar">
                <h2 style="font-size: 14px; margin-bottom: 12px; color: #60a5fa; text-transform: uppercase;">📋 Tasks</h2>
                <div id="tasksList"></div>
                <div style="background: #0f172a; border: 1px solid #334155; border-radius: 6px; padding: 12px; margin-top: 15px;">
                    <h3 style="font-size: 12px; color: #60a5fa; text-transform: uppercase; margin-bottom: 8px;">📊 Stats</h3>
                    <div id="stats" style="display: grid; grid-template-columns: 1fr 1fr; gap: 8px;">
                        <div style="background: #0f172a; border: 1px solid #334155; border-radius: 4px; padding: 8px; text-align: center;"><div style="font-size: 16px; font-weight: 600; color: #667eea;">-</div><div style="font-size: 10px; color: #94a3b8;">Tasks</div></div>
                        <div style="background: #0f172a; border: 1px solid #334155; border-radius: 4px; padding: 8px; text-align: center;"><div style="font-size: 16px; font-weight: 600; color: #667eea;">-</div><div style="font-size: 10px; color: #94a3b8;">Edges</div></div>
                    </div>
                </div>
            </div>
        </div>
        <div id="argo" class="content hidden">
            <div class="code-viewer">
                <h2 style="margin-bottom: 15px;">Argo Workflows YAML</h2>
                <div id="argoContent"></div>
            </div>
        </div>
        <div id="airflow" class="content hidden">
            <div class="code-viewer">
                <h2 style="margin-bottom: 15px;">Apache Airflow DAG</h2>
                <div id="airflowContent"></div>
            </div>
        </div>
        <div id="terraform" class="content hidden">
            <div class="code-viewer">
                <h2 style="margin-bottom: 15px;">Terraform Configuration</h2>
                <div id="terraformContent"></div>
            </div>
        </div>
        <div id="execution" class="content hidden">
            <div class="execution-view" style="flex: 1; overflow-y: auto;">
                <div id="executionContent"></div>
            </div>
        </div>
    </div>
    <script>
        let cy, pipelineData;
        
        async function init() {
            await loadPipelineData();
            initDAG();
            loadAllContent();
        }
        
        async function loadPipelineData() {
            const resp = await fetch('/api/pipeline');
            pipelineData = await resp.json();
        }
        
        async function loadAllContent() {
            const argoResp = await fetch('/api/argo');
            const argoData = await argoResp.json();
            document.getElementById('argoContent').innerHTML = '<div class="code-block"><pre>' + JSON.stringify(argoData, null, 2) + '</pre></div>';
            
            const airflowResp = await fetch('/api/airflow');
            const airflowData = await airflowResp.json();
            const airflowCode = airflowData.code_sample || JSON.stringify(airflowData, null, 2);
            document.getElementById('airflowContent').innerHTML = '<div class="code-block"><pre>' + airflowCode + '</pre></div>';
            
            const terraformResp = await fetch('/api/terraform');
            const terraformData = await terraformResp.json();
            document.getElementById('terraformContent').innerHTML = '<div class="code-block"><pre>' + terraformData.hcl_code + '</pre></div>';
            
			const executionResp = await fetch('/api/execution');
			const executionData = await executionResp.json();
			renderExecution(executionData);
			// start polling for live updates
			setInterval(async () => {
				try {
					const r = await fetch('/api/execution');
					const d = await r.json();
					renderExecution(d);
				} catch (e) {
					console.warn('polling /api/execution failed', e);
				}
			}, 5000);
        }
        
        function renderExecution(data) {
            let html = '<div style="margin-bottom: 20px;"><h3 style="color: #60a5fa; margin-bottom: 10px;">Pipeline: ' + data.pipeline_id + ' <span style="color: #667eea; font-size: 12px;">Status: ' + data.status.toUpperCase() + '</span></h3><p style="font-size: 12px; color: #94a3b8;">Started: ' + new Date(data.started_at).toLocaleString() + '</p></div>';
            
            data.tasks.forEach(function(task) {
                html += '<div class="task-status ' + task.status + '">';
                html += '<div class="task-status-header">';
                html += '<span style="font-weight: 600;">' + task.name + '</span>';
                html += '<span class="status-badge badge-' + task.status + '">' + task.status + '</span>';
                if (task.progress) html += '<span style="color: #94a3b8; font-size: 12px;">' + task.progress + '</span>';
                html += '</div>';
                if (task.status === 'running') {
                    html += '<div class="progress-bar"><div class="progress-fill"></div></div>';
                }
                html += '<p style="font-size: 11px; color: #cbd5e1; margin-bottom: 8px;">Duration: ' + task.duration + '</p>';
                html += '<div class="task-logs">' + task.logs.replace(/\n/g, '<br>') + '</div>';
                html += '</div>';
            });
            
            html += '<div style="background: #0f172a; border: 1px solid #334155; border-radius: 6px; padding: 15px; margin-top: 20px;"><h3 style="color: #60a5fa; margin-bottom: 10px;">📊 Metrics</h3><div style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px;">';
            Object.entries(data.metrics).forEach(function([key, val]) {
                html += '<div style="background: #1e293b; padding: 10px; border-radius: 4px; text-align: center;"><div style="color: #667eea; font-weight: 600;">' + val + '</div><div style="font-size: 11px; color: #94a3b8;">' + key.replace(/_/g, ' ') + '</div></div>';
            });
            html += '</div></div>';
            
            document.getElementById('executionContent').innerHTML = html;
        }
        
        function initDAG() {
            const elements = [];
            Object.entries(pipelineData.tasks).forEach(function([name, task]) {
                elements.push({data: {id: name, label: name, type: task.type, desc: task.description, data: task}});
            });
            pipelineData.edges.forEach(function(e, i) {
                elements.push({data: {id: 'e' + i, source: e.from.task, target: e.to.task}});
            });
            cy = cytoscape({
                container: document.getElementById('cy'),
                elements: elements,
                style: [
                    {selector: 'node[type="Source"]', style: {'background-color': '#10b981', 'label': 'data(label)', 'text-valign': 'center', 'text-halign': 'center', 'color': '#fff', 'font-size': '13px', 'font-weight': 'bold', 'width': '100px', 'height': '100px', 'border-width': '2px', 'border-color': '#1e293b'}},
                    {selector: 'node[type="Transform"]', style: {'background-color': '#667eea', 'label': 'data(label)', 'text-valign': 'center', 'text-halign': 'center', 'color': '#fff', 'font-size': '13px', 'font-weight': 'bold', 'width': '100px', 'height': '100px', 'border-width': '2px', 'border-color': '#1e293b'}},
                    {selector: 'node[type="Sink"]', style: {'background-color': '#f59e0b', 'label': 'data(label)', 'text-valign': 'center', 'text-halign': 'center', 'color': '#fff', 'font-size': '13px', 'font-weight': 'bold', 'width': '100px', 'height': '100px', 'border-width': '2px', 'border-color': '#1e293b'}},
                    {selector: 'edge', style: {'target-arrow-shape': 'triangle', 'target-arrow-color': '#94a3b8', 'line-color': '#64748b', 'width': '2px'}}
                ],
                layout: {name: 'breadthfirst', directed: true, spacingFactor: 2}
            });
            renderList();
            updateStats();
        }
        
        function renderList() {
            const list = document.getElementById('tasksList');
            list.innerHTML = '';
            Object.entries(pipelineData.tasks).forEach(function([name, task]) {
                const card = document.createElement('div');
                card.style.cssText = 'background: #0f172a; border: 1px solid #334155; border-radius: 6px; padding: 10px; margin-bottom: 10px; cursor: pointer; transition: all 0.2s ease;';
                card.innerHTML = '<div style="font-size: 10px; color: #94a3b8; margin-bottom: 4px;">' + task.type + '</div><div style="font-weight: 600; font-size: 13px; margin-bottom: 4px;">' + name + '</div><div style="font-size: 11px; color: #cbd5e1;">' + task.description + '</div>';
                card.onmouseover = () => card.style.borderColor = '#667eea';
                card.onmouseout = () => card.style.borderColor = '#334155';
                list.appendChild(card);
            });
        }
        
        function updateStats() {
            const st = document.getElementById('stats');
            st.innerHTML = '<div style="background: #0f172a; border: 1px solid #334155; border-radius: 4px; padding: 8px; text-align: center;"><div style="font-size: 16px; font-weight: 600; color: #667eea;">' + Object.keys(pipelineData.tasks).length + '</div><div style="font-size: 10px; color: #94a3b8;">Tasks</div></div><div style="background: #0f172a; border: 1px solid #334155; border-radius: 4px; padding: 8px; text-align: center;"><div style="font-size: 16px; font-weight: 600; color: #667eea;">' + pipelineData.edges.length + '</div><div style="font-size: 10px; color: #94a3b8;">Edges</div></div>';
        }
        
        function switchTab(tab) {
            document.querySelectorAll('.content').forEach(c => c.classList.add('hidden'));
            document.getElementById(tab).classList.remove('hidden');
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            event.target.classList.add('active');
        }
        
        window.onload = init;
    </script>
</body>
</html>`
}
