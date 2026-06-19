package airflow

import (
	"context"
	"fmt"
	"strings"

	"flowforge/compiler/pkg"
	ir "flowforge/ir/pkg"
)

// AirflowCompiler compiles PipelineSpec to Apache Airflow DAG Python code
type AirflowCompiler struct {
	dagID string
}

// New creates a new Airflow compiler
func New(dagID string) *AirflowCompiler {
	if dagID == "" {
		dagID = "flowforge_dag"
	}
	return &AirflowCompiler{dagID: dagID}
}

// Compile implements ExecutorCompiler
func (c *AirflowCompiler) Compile(
	ctx context.Context,
	spec *ir.PipelineSpec,
) (pkg.CompileResult, error) {
	builder := NewBuilder(spec, c.dagID)

	// Add all tasks
	for taskID, task := range spec.Tasks {
		if err := builder.AddTask(taskID, task); err != nil {
			return pkg.CompileResult{}, fmt.Errorf("failed to add task %s: %w", taskID, err)
		}
	}

	// Add edges
	for _, edge := range spec.Edges {
		builder.AddEdge(edge.From, edge.To)
	}

	// Generate Python code
	code, err := builder.Build(ctx)
	if err != nil {
		return pkg.CompileResult{}, fmt.Errorf("failed to build DAG: %w", err)
	}

	return pkg.CompileResult{
		Format:   pkg.ExecutorFormatAirflow,
		Artifact: code,
		Metadata: map[string]interface{}{
			"dag_id":      c.dagID,
			"task_count":  len(spec.Tasks),
			"edge_count":  len(spec.Edges),
			"pipeline_id": spec.Metadata.Name,
		},
	}, nil
}

// GetFormat implements ExecutorCompiler
func (c *AirflowCompiler) GetFormat() pkg.ExecutorFormat {
	return pkg.ExecutorFormatAirflow
}

// Task represents an Airflow task
type Task struct {
	ID           string
	OperatorType string
	ImageName    string
	Command      []string
	Environment  map[string]string
	Dependencies []string
}

// Builder builds Airflow DAGs
type Builder struct {
	spec  *ir.PipelineSpec
	dagID string
	tasks map[string]*Task
	edges []struct{ From, To string }
}

// NewBuilder creates a new builder
func NewBuilder(spec *ir.PipelineSpec, dagID string) *Builder {
	return &Builder{
		spec:  spec,
		dagID: dagID,
		tasks: make(map[string]*Task),
	}
}

// AddTask adds a task to the DAG
func (b *Builder) AddTask(taskID string, task *ir.Task) error {
	t := &Task{
		ID:           taskID,
		OperatorType: "KubernetesPodOperator", // Default to Kubernetes operator
		ImageName:    task.Config.Image,
		Command:      task.Config.Command,
		Environment:  make(map[string]string),
	}

	// Convert environment
	for k, v := range task.Config.Env {
		if str, ok := v.(string); ok {
			t.Environment[k] = str
		}
	}

	// Determine operator type based on handler
	if task.Handler != nil {
		switch task.Handler.Type {
		case "python":
			t.OperatorType = "PythonOperator"
		case "bash":
			t.OperatorType = "BashOperator"
		case "spark":
			t.OperatorType = "SparkSubmitOperator"
		default:
			t.OperatorType = "KubernetesPodOperator"
		}
	}

	b.tasks[taskID] = t
	return nil
}

// AddEdge adds a dependency edge
func (b *Builder) AddEdge(from, to string) {
	b.edges = append(b.edges, struct{ From, To string }{From: from, To: to})

	// Record dependency in task
	if task, exists := b.tasks[to]; exists {
		task.Dependencies = append(task.Dependencies, from)
	}
}

// Build generates the Python DAG code
func (b *Builder) Build(ctx context.Context) (string, error) {
	code := b.generateImports()
	code += b.generateDAGDefinition()
	code += b.generateTasks()
	code += b.generateDependencies()

	return code, nil
}

// generateImports generates import statements
func (b *Builder) generateImports() string {
	code := "\"\"\"FlowForge-generated Airflow DAG\"\"\"\n\n"
	code += "from datetime import datetime, timedelta\n"
	code += "from airflow import DAG\n"
	code += "from airflow.operators.python import PythonOperator\n"
	code += "from airflow.operators.bash import BashOperator\n"
	code += "from airflow.providers.cncf.kubernetes.operators.kubernetes_pod import KubernetesPodOperator\n"
	code += "from airflow.models import Variable\n\n"
	return code
}

// generateDAGDefinition generates the DAG definition
func (b *Builder) generateDAGDefinition() string {
	code := "# DAG Configuration\n"
	code += fmt.Sprintf("dag_id = '%s'\n", b.dagID)
	code += fmt.Sprintf("pipeline_name = '%s'\n", b.spec.Metadata.Name)
	code += fmt.Sprintf("pipeline_version = '%s'\n", b.spec.Metadata.Version)
	code += fmt.Sprintf("pipeline_owner = '%s'\n\n", b.spec.Metadata.Owner)

	code += "default_args = {\n"
	code += "    'owner': pipeline_owner,\n"
	code += "    'retries': 1,\n"
	code += "    'retry_delay': timedelta(minutes=5),\n"
	code += "    'start_date': datetime(2024, 1, 1),\n"
	code += "}\n\n"

	code += "dag = DAG(\n"
	code += fmt.Sprintf("    dag_id='%s',\n", b.dagID)
	code += fmt.Sprintf("    description='FlowForge pipeline: %s',\n", b.spec.Metadata.Name)
	code += "    default_args=default_args,\n"
	code += "    schedule_interval=None,\n"
	code += "    catchup=False,\n"
	code += "    tags=['flowforge', '" + b.spec.Metadata.Name + "'],\n"
	code += ")\n\n"

	return code
}

// generateTasks generates task definitions
func (b *Builder) generateTasks() string {
	code := "# Task Definitions\n"

	for taskID, task := range b.tasks {
		code += b.generateTaskDefinition(taskID, task)
	}

	code += "\n"
	return code
}

// generateTaskDefinition generates a single task definition
func (b *Builder) generateTaskDefinition(taskID string, task *Task) string {
	code := fmt.Sprintf("# Task: %s\n", taskID)
	code += fmt.Sprintf("%s = KubernetesPodOperator(\n", sanitizeTaskName(taskID))
	code += fmt.Sprintf("    task_id='%s',\n", taskID)
	code += fmt.Sprintf("    image='%s',\n", task.ImageName)

	if len(task.Command) > 0 {
		code += fmt.Sprintf("    cmds=%v,\n", task.Command)
	}

	if len(task.Environment) > 0 {
		code += "    env_vars={\n"
		for k, v := range task.Environment {
			code += fmt.Sprintf("        '%s': '%s',\n", k, v)
		}
		code += "    },\n"
	}

	code += "    namespace='default',\n"
	code += "    dag=dag,\n"
	code += ")\n\n"

	return code
}

// generateDependencies generates task dependencies
func (b *Builder) generateDependencies() string {
	if len(b.edges) == 0 {
		return ""
	}

	code := "# Task Dependencies\n"

	for _, edge := range b.edges {
		fromTask := sanitizeTaskName(edge.From)
		toTask := sanitizeTaskName(edge.To)
		code += fmt.Sprintf("%s >> %s\n", fromTask, toTask)
	}

	return code
}

// sanitizeTaskName converts a task ID to a valid Python variable name
func sanitizeTaskName(taskID string) string {
	// Replace invalid characters with underscores
	name := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, taskID)

	// Ensure it doesn't start with a digit
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "t_" + name
	}

	return name
}
