import { PipelineSpec } from '../parser/pythonParser';

function timeoutToSeconds(timeout?: string): number | undefined {
  if (!timeout) {
    return undefined;
  }
  const m = /^(\d+)s$/.exec(timeout.trim());
  if (!m) {
    return undefined;
  }
  return Number(m[1]);
}

export function compileToArgo(spec: PipelineSpec): string {
  const depsByTarget = new Map<string, string[]>();
  for (const edge of spec.edges) {
    if (!depsByTarget.has(edge.to)) {
      depsByTarget.set(edge.to, []);
    }
    depsByTarget.get(edge.to)!.push(edge.from);
  }

  const lines: string[] = [];
  lines.push('apiVersion: argoproj.io/v1alpha1');
  lines.push('kind: Workflow');
  lines.push('metadata:');
  lines.push(`  name: ${spec.meta.name}`);
  lines.push('  namespace: default');
  lines.push('spec:');
  const firstTimeout = timeoutToSeconds(spec.tasks.find(t => t.timeout)?.timeout);
  if (firstTimeout) {
    lines.push(`  activeDeadlineSeconds: ${firstTimeout}`);
  }
  lines.push(`  entrypoint: ${spec.meta.name}`);
  lines.push('  templates:');
  lines.push(`  - name: ${spec.meta.name}`);
  lines.push('    dag:');
  lines.push('      tasks:');

  for (const task of spec.tasks) {
    lines.push(`      - name: ${task.id}`);
    lines.push(`        template: ${task.id}`);
    const deps = depsByTarget.get(task.id) || [];
    if (deps.length > 0) {
      lines.push(`        dependencies: [${deps.join(', ')}]`);
    }
  }

  for (const task of spec.tasks) {
    lines.push(`  - name: ${task.id}`);
    lines.push('    container:');
    lines.push(`      image: ${task.image}`);
    lines.push(`      command: ["python", "${task.id}.py"]`);
    lines.push('      resources:');
    lines.push('        requests:');
    lines.push('          memory: "512Mi"');
    lines.push('          cpu: "100m"');
    if (typeof task.retries === 'number') {
      lines.push('    retryStrategy:');
      lines.push(`      limit: ${task.retries}`);
    }
  }

  return lines.join('\n') + '\n';
}
