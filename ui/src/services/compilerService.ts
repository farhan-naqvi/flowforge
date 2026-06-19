// ui/src/services/compilerService.ts - Client-side compiler integration

import { Pipeline, CompilationResult, ValidationError } from '../types/flowforge';

export class CompilerService {
  private apiBaseUrl = process.env.REACT_APP_COMPILER_API || 'http://localhost:8080/api';

  // Convert pipeline to IR JSON
  pipelineToIR(pipeline: Pipeline): string {
    return JSON.stringify({
      metadata: {
        name: pipeline.name,
        version: pipeline.version,
        description: pipeline.description,
        ...pipeline.metadata,
      },
      tasks: Object.entries(pipeline.tasks).reduce((acc, [id, task]) => {
        acc[id] = {
          handler: task.handler,
          config: task.config,
        };
        return acc;
      }, {} as Record<string, any>),
      edges: pipeline.edges.map(edge => ({
        from: edge.from,
        to: edge.to,
        fromPort: edge.fromPort || 'output',
        toPort: edge.toPort || 'input',
        condition: edge.condition,
      })),
    }, null, 2);
  }

  // Compile to Argo or Airflow
  async compile(pipeline: Pipeline, executor: 'argo' | 'airflow'): Promise<CompilationResult> {
    const ir = this.pipelineToIR(pipeline);

    const response = await fetch(`${this.apiBaseUrl}/compile`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ir,
        executor,
      }),
    });

    if (!response.ok) {
      throw new Error(`Compilation failed: ${response.statusText}`);
    }

    return response.json();
  }

  // Validate pipeline
  async validate(pipeline: Pipeline): Promise<ValidationError[]> {
    const ir = this.pipelineToIR(pipeline);

    const response = await fetch(`${this.apiBaseUrl}/validate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ir }),
    });

    if (!response.ok) {
      throw new Error(`Validation failed: ${response.statusText}`);
    }

    const result = await response.json();
    return result.errors || [];
  }

  // Get compilation estimate
  async getEstimate(pipeline: Pipeline): Promise<{ cost: number; time: number }> {
    const ir = this.pipelineToIR(pipeline);

    const response = await fetch(`${this.apiBaseUrl}/estimate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ir }),
    });

    if (!response.ok) {
      throw new Error(`Estimation failed: ${response.statusText}`);
    }

    return response.json();
  }

  // Export pipeline
  exportPipeline(pipeline: Pipeline, format: 'json' | 'yaml' | 'py'): string {
    const ir = this.pipelineToIR(pipeline);

    if (format === 'json') {
      return ir;
    }

    if (format === 'yaml') {
      return this.jsonToYAML(JSON.parse(ir));
    }

    if (format === 'py') {
      return this.jsonToPython(JSON.parse(ir));
    }

    return ir;
  }

  // Import pipeline from JSON/YAML/Python
  async importPipeline(content: string, format: 'json' | 'yaml' | 'py'): Promise<Pipeline> {
    let ir: any;

    if (format === 'json') {
      ir = JSON.parse(content);
    } else if (format === 'yaml') {
      ir = this.yamlToJSON(content);
    } else if (format === 'py') {
      ir = this.pythonToJSON(content);
    }

    return this.irToPipeline(ir);
  }

  private irToPipeline(ir: any): Pipeline {
    return {
      id: ir.metadata.name,
      name: ir.metadata.name,
      version: ir.metadata.version || '1.0.0',
      description: ir.metadata.description || '',
      tasks: Object.entries(ir.tasks || {}).reduce((acc, [id, task]: [string, any]) => {
        acc[id] = {
          id,
          name: id,
          handler: task.handler,
          config: task.config,
          inputs: [],
          outputs: [],
          dependencies: [],
        };
        return acc;
      }, {} as Record<string, Task>),
      edges: (ir.edges || []).map((e: any) => ({
        from: e.from,
        to: e.to,
        fromPort: e.fromPort,
        toPort: e.toPort,
        condition: e.condition,
      })),
      metadata: ir.metadata,
      createdAt: new Date(),
      updatedAt: new Date(),
    };
  }

  private jsonToYAML(obj: any): string {
    // Simple JSON to YAML conversion
    return this.objToYAML(obj);
  }

  private yamlToJSON(yaml: string): any {
    // Simple YAML to JSON conversion (would use yaml-js library in real app)
    return JSON.parse(yaml);
  }

  private pythonToJSON(py: string): any {
    // Parse Python decorator-based pipeline definition
    // This is a simplified version - real implementation would be more complex
    const metadataMatch = py.match(/@pipeline\(name="([^"]+)"/);
    return {
      metadata: {
        name: metadataMatch?.[1] || 'pipeline',
      },
      tasks: {},
      edges: [],
    };
  }

  private jsonToPython(ir: any): string {
    let py = 'from flowforge import pipeline, task\n\n';
    py += `@pipeline(name="${ir.metadata.name}", version="${ir.metadata.version}")\n`;
    py += 'def ' + ir.metadata.name.replace(/-/g, '_') + '():\n';
    py += '    pass\n\n';

    for (const [taskId, task] of Object.entries(ir.tasks || {})) {
      py += `@task(image="${task.config.image}")\n`;
      py += `def ${taskId}():\n`;
      py += `    # ${task.handler.command}\n`;
      py += '    pass\n\n';
    }

    return py;
  }

  private objToYAML(obj: any, indent = 0): string {
    let yaml = '';
    const indentStr = ' '.repeat(indent);

    if (Array.isArray(obj)) {
      for (const item of obj) {
        yaml += indentStr + '- ' + this.objToYAML(item, indent + 2).trim() + '\n';
      }
    } else if (typeof obj === 'object' && obj !== null) {
      for (const [key, value] of Object.entries(obj)) {
        if (typeof value === 'object' && value !== null) {
          yaml += indentStr + key + ':\n' + this.objToYAML(value, indent + 2);
        } else {
          yaml += indentStr + key + ': ' + JSON.stringify(value) + '\n';
        }
      }
    } else {
      yaml = JSON.stringify(obj);
    }

    return yaml;
  }
}

// Import Task type
import { Task } from '../types/flowforge';

export const compilerService = new CompilerService();
