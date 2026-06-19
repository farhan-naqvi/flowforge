package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"flowforge/ir"
)

// TerraformGenerator generates Terraform configurations from IR specs
type TerraformGenerator struct {
	provider string
}

// NewTerraformGenerator creates a new Terraform generator
func NewTerraformGenerator(provider string) *TerraformGenerator {
	return &TerraformGenerator{
		provider: provider,
	}
}

// Generate creates Terraform configuration
func (g *TerraformGenerator) Generate(ctx context.Context, spec *ir.PipelineSpec) (string, error) {
	tmpl := `
# Generated Terraform configuration for FlowForge pipeline
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "{{ .Provider }}" {
  # Configure your provider here
}

{{ .ResourceDefinitions }}

{{ .OutputDefinitions }}
`

	t, err := template.New("terraform").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("template parse failed: %w", err)
	}

	// Generate resource definitions
	resources := g.generateResources(spec)
	outputs := g.generateOutputs(spec)

	data := map[string]interface{}{
		"Provider":            g.provider,
		"ResourceDefinitions": resources,
		"OutputDefinitions":   outputs,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

// generateResources creates resource blocks
func (g *TerraformGenerator) generateResources(spec *ir.PipelineSpec) string {
	var buf bytes.Buffer

	pipelineName := spec.Metadata["name"].(string)

	// Namespace resource
	buf.WriteString(fmt.Sprintf(`
resource "kubernetes_namespace" "pipeline" {
  metadata {
    name = "%s"
  }
}
`, pipelineName))

	// Task deployment resources
	for taskID, task := range spec.Tasks {
		buf.WriteString(fmt.Sprintf(`
resource "kubernetes_deployment" "%s" {
  metadata {
    name      = "%s"
    namespace = kubernetes_namespace.pipeline.metadata[0].name
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "%s"
      }
    }

    template {
      metadata {
        labels = {
          app = "%s"
        }
      }

      spec {
        container {
          name  = "%s"
          image = "%s"
        }
      }
    }
  }
}
`, taskID, taskID, taskID, taskID, taskID, task.Config.Image))
	}

	return buf.String()
}

// generateOutputs creates output blocks
func (g *TerraformGenerator) generateOutputs(spec *ir.PipelineSpec) string {
	var buf bytes.Buffer

	pipelineName := spec.Metadata["name"].(string)

	buf.WriteString(fmt.Sprintf(`
output "namespace" {
  value = kubernetes_namespace.pipeline.metadata[0].name
}

output "deployment_status" {
  value = "Pipeline %s deployed"
}
`, pipelineName))

	return buf.String()
}

// HelmGenerator generates Helm chart configurations from IR specs
type HelmGenerator struct {
	chartName    string
	chartVersion string
}

// NewHelmChartGenerator creates a new Helm chart generator
func NewHelmChartGenerator(chartName, chartVersion string) *HelmGenerator {
	return &HelmGenerator{
		chartName:    chartName,
		chartVersion: chartVersion,
	}
}

// GenerateChart creates a complete Helm chart structure
func (g *HelmChartGenerator) GenerateChart(ctx context.Context, spec *ir.PipelineSpec) (map[string]interface{}, error) {
	pipelineName := spec.Metadata["name"].(string)
	pipelineVersion := spec.Metadata["version"].(string)

	chart := map[string]interface{}{
		"apiVersion":  "v2",
		"name":        g.chartName,
		"version":     g.chartVersion,
		"appVersion":  pipelineVersion,
		"description": "Helm chart for FlowForge pipeline",
		"type":        "application",
		"home":        "https://flowforge.dev",
		"sources":     []string{"https://github.com/flowforge/flowforge"},
		"maintainers": []map[string]string{
			{
				"name":  "FlowForge Team",
				"email": "team@flowforge.dev",
			},
		},
	}

	return chart, nil
}

// GenerateValues creates default Helm values
func (g *HelmChartGenerator) GenerateValues(ctx context.Context, spec *ir.PipelineSpec) (map[string]interface{}, error) {
	pipelineName := spec.Metadata["name"].(string)

	values := map[string]interface{}{
		"replicaCount": 1,
		"namespace":    pipelineName,
		"image": map[string]interface{}{
			"pullPolicy": "IfNotPresent",
		},
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "1000m",
				"memory": "512Mi",
			},
			"requests": map[string]interface{}{
				"cpu":    "100m",
				"memory": "128Mi",
			},
		},
		"services": g.generateServiceValues(spec),
		"tasks":    g.generateTaskValues(spec),
	}

	return values, nil
}

// generateServiceValues creates service configurations
func (g *HelmChartGenerator) generateServiceValues(spec *ir.PipelineSpec) []map[string]interface{} {
	services := make([]map[string]interface{}, 0)

	pipelineName := spec.Metadata["name"].(string)

	services = append(services, map[string]interface{}{
		"name":       pipelineName,
		"type":       "ClusterIP",
		"port":       80,
		"targetPort": 8080,
	})

	return services
}

// generateTaskValues creates task configurations
func (g *HelmChartGenerator) generateTaskValues(spec *ir.PipelineSpec) []map[string]interface{} {
	tasks := make([]map[string]interface{}, 0)

	for taskID, task := range spec.Tasks {
		taskConfig := map[string]interface{}{
			"name":    taskID,
			"image":   task.Config.Image,
			"command": task.Handler.Command,
		}

		if task.Config.Resources != nil {
			taskConfig["resources"] = task.Config.Resources
		}

		if task.Config.Timeout > 0 {
			taskConfig["timeout"] = task.Config.Timeout
		}

		tasks = append(tasks, taskConfig)
	}

	return tasks
}

// GenerateDeployment creates a Helm deployment template
func (g *HelmChartGenerator) GenerateDeployment(ctx context.Context, spec *ir.PipelineSpec) (string, error) {
	tmpl := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "chart.fullname" . }}
  namespace: {{ .Values.namespace }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: {{ include "chart.name" . }}
  template:
    metadata:
      labels:
        app: {{ include "chart.name" . }}
    spec:
      containers:
      {{- range .Values.tasks }}
      - name: {{ .name }}
        image: {{ .image }}
        resources: {{ toJson .resources | nindent 12 }}
      {{- end }}
`
	return tmpl, nil
}

// PlanEstimator estimates deployment impact
type PlanEstimator struct{}

// EstimateImpact estimates the impact of a deployment
func (pe *PlanEstimator) EstimateImpact(ctx context.Context, spec *ir.PipelineSpec) (*PlanEstimate, error) {
	estimate := &PlanEstimate{
		ResourcesAdded: len(spec.Tasks) + 1, // Tasks + namespace
		EstimatedCost:  float64(len(spec.Tasks)) * 0.05,
		EstimatedTime:  5 * 60, // 5 minutes in seconds
	}

	return estimate, nil
}

// ConfigGenerator generates complete deployment configurations
type ConfigGenerator struct {
	tfGen     *TerraformGenerator
	helmGen   *HelmChartGenerator
	estimator *PlanEstimator
}

// NewConfigGenerator creates a new configuration generator
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{
		tfGen:     NewTerraformGenerator("kubernetes"),
		helmGen:   NewHelmChartGenerator("flowforge-pipeline", "1.0.0"),
		estimator: &PlanEstimator{},
	}
}

// GenerateDeploymentConfig generates all deployment configurations
func (cg *ConfigGenerator) GenerateDeploymentConfig(ctx context.Context, spec *ir.PipelineSpec) (*DeploymentConfig, error) {
	// Generate Terraform
	terraform, err := cg.tfGen.Generate(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("terraform generation failed: %w", err)
	}

	// Generate Helm chart
	chart, err := cg.helmGen.GenerateChart(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("helm chart generation failed: %w", err)
	}

	// Generate Helm values
	values, err := cg.helmGen.GenerateValues(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("helm values generation failed: %w", err)
	}

	// Estimate impact
	estimate, err := cg.estimator.EstimateImpact(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("impact estimation failed: %w", err)
	}

	config := &DeploymentConfig{
		PipelineSpec: spec,
		Terraform:    terraform,
		HelmChart:    chart,
		HelmValues:   values,
		Estimate:     estimate,
		GeneratedAt:  time.Now(),
	}

	return config, nil
}

// DeploymentConfig contains all generated deployment configurations
type DeploymentConfig struct {
	PipelineSpec *ir.PipelineSpec
	Terraform    string
	HelmChart    map[string]interface{}
	HelmValues   map[string]interface{}
	Estimate     *PlanEstimate
	GeneratedAt  time.Time
}

// String returns JSON representation
func (dc *DeploymentConfig) String() string {
	data, _ := json.MarshalIndent(dc, "", "  ")
	return string(data)
}
