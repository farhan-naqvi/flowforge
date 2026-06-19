package argo

import (
	"context"
	"encoding/json"
	"fmt"

	"flowforge/compiler/pkg"
	"flowforge/ir/pkg"
	ir "flowforge/ir/pkg"
)

// ArgoCompiler compiles PipelineSpec to Argo Workflows YAML
type ArgoCompiler struct {
	namespace string
}

// New creates a new Argo compiler
func New(namespace string) *ArgoCompiler {
	if namespace == "" {
		namespace = "default"
	}
	return &ArgoCompiler{namespace: namespace}
}

// Compile implements ExecutorCompiler
func (c *ArgoCompiler) Compile(
	ctx context.Context,
	spec *ir.PipelineSpec,
) (pkg.CompileResult, error) {
	builder := NewBuilder(spec, c.namespace)

	// Add all tasks as templates
	for taskID, task := range spec.Tasks {
		if err := builder.AddTask(taskID, task); err != nil {
			return pkg.CompileResult{}, fmt.Errorf("failed to add task %s: %w", taskID, err)
		}
	}

	// Build workflow
	workflow, err := builder.Build(ctx)
	if err != nil {
		return pkg.CompileResult{}, fmt.Errorf("failed to build workflow: %w", err)
	}

	// Generate YAML
	yamlStr, err := builder.ToYAML(workflow)
	if err != nil {
		return pkg.CompileResult{}, fmt.Errorf("failed to generate YAML: %w", err)
	}

	return pkg.CompileResult{
		Format:   pkg.ExecutorFormatArgo,
		Artifact: yamlStr,
		Metadata: map[string]interface{}{
			"namespace":   c.namespace,
			"task_count":  len(spec.Tasks),
			"edge_count":  len(spec.Edges),
			"pipeline_id": spec.Metadata.Name,
		},
	}, nil
}

// GetFormat implements ExecutorCompiler
func (c *ArgoCompiler) GetFormat() pkg.ExecutorFormat {
	return pkg.ExecutorFormatArgo
}

// ArgoWorkflow represents an Argo Workflow
type ArgoWorkflow struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       *ArgoWorkflowSpec      `json:"spec"`
}

// ArgoWorkflowSpec is the spec section
type ArgoWorkflowSpec struct {
	Entrypoint string         `json:"entrypoint"`
	Templates  []ArgoTemplate `json:"templates"`
	Arguments  *ArgoArguments `json:"arguments,omitempty"`
}

// ArgoTemplate represents a template
type ArgoTemplate struct {
	Name    string               `json:"name"`
	Inputs  *ArgoTemplateInputs  `json:"inputs,omitempty"`
	Outputs *ArgoTemplateOutputs `json:"outputs,omitempty"`

	// Container task
	Container *ArgoContainer `json:"container,omitempty"`

	// DAG task
	DAG *ArgoDAG `json:"dag,omitempty"`
}

// ArgoContainer represents a container spec
type ArgoContainer struct {
	Image     string         `json:"image"`
	Command   []string       `json:"command,omitempty"`
	Args      []string       `json:"args,omitempty"`
	Env       []ArgoEnvVar   `json:"env,omitempty"`
	Resources *ArgoResources `json:"resources,omitempty"`
}

// ArgoEnvVar represents an environment variable
type ArgoEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ArgoResources represents resource requests/limits
type ArgoResources struct {
	Requests map[string]string `json:"requests,omitempty"`
	Limits   map[string]string `json:"limits,omitempty"`
}

// ArgoDAG represents a DAG template
type ArgoDAG struct {
	Tasks []ArgoTask `json:"tasks"`
}

// ArgoTask represents a task in a DAG
type ArgoTask struct {
	Name         string                 `json:"name"`
	Template     string                 `json:"template"`
	Dependencies string                 `json:"dependencies,omitempty"`
	Arguments    map[string]interface{} `json:"arguments,omitempty"`
	When         string                 `json:"when,omitempty"`
}

// ArgoTemplateInputs represents inputs
type ArgoTemplateInputs struct {
	Parameters []ArgoParameter `json:"parameters,omitempty"`
}

// ArgoTemplateOutputs represents outputs
type ArgoTemplateOutputs struct {
	Parameters []ArgoParameter `json:"parameters,omitempty"`
	Result     *ArgoResult     `json:"result,omitempty"`
}

// ArgoParameter represents a parameter
type ArgoParameter struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
	Path  string `json:"path,omitempty"`
}

// ArgoResult represents a result output
type ArgoResult struct {
	Path string `json:"path"`
}

// ArgoArguments represents arguments
type ArgoArguments struct {
	Parameters []ArgoParameter `json:"parameters,omitempty"`
}

// Builder builds Argo workflows
type Builder struct {
	spec      *ir.PipelineSpec
	namespace string
	templates map[string]*ArgoTemplate
	dagTasks  []ArgoTask
}

// NewBuilder creates a new builder
func NewBuilder(spec *ir.PipelineSpec, namespace string) *Builder {
	return &Builder{
		spec:      spec,
		namespace: namespace,
		templates: make(map[string]*ArgoTemplate),
		dagTasks:  []ArgoTask{},
	}
}

// AddTask adds a task to the workflow
func (b *Builder) AddTask(taskID string, task *ir.Task) error {
	// Create container task template
	template := &ArgoTemplate{
		Name: taskID,
		Container: &ArgoContainer{
			Image:   task.Config.Image,
			Command: task.Config.Command,
		},
	}

	// Add environment variables
	if len(task.Config.Env) > 0 {
		envVars := []ArgoEnvVar{}
		for k, v := range task.Config.Env {
			envVars = append(envVars, ArgoEnvVar{
				Name:  k,
				Value: v.(string),
			})
		}
		template.Container.Env = envVars
	}

	// Add resources
	if task.Config.Resources != nil {
		template.Container.Resources = &ArgoResources{
			Requests: make(map[string]string),
			Limits:   make(map[string]string),
		}
		// Parse resources from config
	}

	b.templates[taskID] = template

	// Add to DAG task list
	dagTask := ArgoTask{
		Name:     taskID,
		Template: taskID,
	}

	// Find dependencies
	for _, edge := range b.spec.Edges {
		if edge.To == taskID {
			if dagTask.Dependencies == "" {
				dagTask.Dependencies = edge.From
			} else {
				dagTask.Dependencies += ", " + edge.From
			}
		}
	}

	b.dagTasks = append(b.dagTasks, dagTask)

	return nil
}

// Build creates the final Argo Workflow
func (b *Builder) Build(ctx context.Context) (*ArgoWorkflow, error) {
	// Create main DAG template
	mainTemplate := &ArgoTemplate{
		Name: b.spec.Metadata.Name,
		DAG: &ArgoDAG{
			Tasks: b.dagTasks,
		},
	}

	// Collect all templates
	allTemplates := []ArgoTemplate{*mainTemplate}
	for _, template := range b.templates {
		allTemplates = append(allTemplates, *template)
	}

	// Create workflow
	workflow := &ArgoWorkflow{
		APIVersion: "argoproj.io/v1alpha1",
		Kind:       "Workflow",
		Metadata: map[string]interface{}{
			"name":      b.spec.Metadata.Name,
			"namespace": b.namespace,
		},
		Spec: &ArgoWorkflowSpec{
			Entrypoint: b.spec.Metadata.Name,
			Templates:  allTemplates,
		},
	}

	return workflow, nil
}

// ToYAML converts to YAML string
func (b *Builder) ToYAML(workflow *ArgoWorkflow) (string, error) {
	// Convert to JSON first, then format as YAML
	data, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return "", err
	}

	// Simple YAML conversion (in production, use gopkg.in/yaml.v3)
	yamlStr := "apiVersion: argoproj.io/v1alpha1\n"
	yamlStr += "kind: Workflow\n"
	yamlStr += fmt.Sprintf("metadata:\n  name: %s\n  namespace: %s\n", workflow.Metadata["name"], workflow.Metadata["namespace"])
	yamlStr += "spec:\n"
	yamlStr += fmt.Sprintf("  entrypoint: %s\n", workflow.Spec.Entrypoint)
	yamlStr += "  templates:\n"

	for _, template := range workflow.Spec.Templates {
		yamlStr += fmt.Sprintf("  - name: %s\n", template.Name)

		if template.Container != nil {
			yamlStr += fmt.Sprintf("    container:\n")
			yamlStr += fmt.Sprintf("      image: %s\n", template.Container.Image)
			if len(template.Container.Command) > 0 {
				yamlStr += fmt.Sprintf("      command: %v\n", template.Container.Command)
			}
		}

		if template.DAG != nil {
			yamlStr += "    dag:\n"
			yamlStr += "      tasks:\n"
			for _, task := range template.DAG.Tasks {
				yamlStr += fmt.Sprintf("      - name: %s\n", task.Name)
				yamlStr += fmt.Sprintf("        template: %s\n", task.Template)
				if task.Dependencies != "" {
					yamlStr += fmt.Sprintf("        dependencies: %s\n", task.Dependencies)
				}
			}
		}
	}

	return yamlStr, nil
}
