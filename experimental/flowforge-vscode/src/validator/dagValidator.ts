import { EdgeSpec, PipelineSpec } from '../parser/pythonParser';

export interface ValidationIssue {
  message: string;
}

function buildAdjacency(edges: EdgeSpec[]): Map<string, string[]> {
  const adj = new Map<string, string[]>();
  for (const edge of edges) {
    if (!adj.has(edge.from)) {
      adj.set(edge.from, []);
    }
    adj.get(edge.from)!.push(edge.to);
  }
  return adj;
}

export function validatePipeline(spec: PipelineSpec): ValidationIssue[] {
  const issues: ValidationIssue[] = [];
  const taskIds = new Set(spec.tasks.map(t => t.id));

  if (taskIds.size !== spec.tasks.length) {
    issues.push({ message: 'Duplicate task IDs are not allowed.' });
  }

  if (!spec.meta.name) {
    issues.push({ message: 'Pipeline name is required.' });
  }

  if (taskIds.size === 0) {
    issues.push({ message: 'Pipeline must define at least one @task.' });
  }

  for (const edge of spec.edges) {
    if (!taskIds.has(edge.from)) {
      issues.push({ message: `Edge source task not found: ${edge.from}` });
    }
    if (!taskIds.has(edge.to)) {
      issues.push({ message: `Edge destination task not found: ${edge.to}` });
    }
  }

  // Cycle detection (DFS colors)
  const adj = buildAdjacency(spec.edges);
  const color = new Map<string, number>(); // 0 white, 1 gray, 2 black

  const visit = (node: string): boolean => {
    color.set(node, 1);
    for (const next of adj.get(node) || []) {
      const c = color.get(next) || 0;
      if (c === 1) {
        return true;
      }
      if (c === 0 && visit(next)) {
        return true;
      }
    }
    color.set(node, 2);
    return false;
  };

  for (const id of taskIds) {
    if ((color.get(id) || 0) === 0) {
      if (visit(id)) {
        issues.push({ message: 'Pipeline contains a cycle.' });
        break;
      }
    }
  }

  return issues;
}
