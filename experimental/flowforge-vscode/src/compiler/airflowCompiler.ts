import { PipelineSpec } from '../parser/pythonParser';

export function compileToAirflow(spec: PipelineSpec): string {
  const lines: string[] = [];
  lines.push('from airflow import DAG');
  lines.push('from airflow.providers.cncf.kubernetes.operators.kubernetes_pod import KubernetesPodOperator');
  lines.push('from datetime import datetime, timedelta');
  lines.push('');
  lines.push('dag = DAG(');
  lines.push(`    dag_id='${spec.meta.name}',`);
  lines.push(`    default_args={'owner': '${spec.meta.owner || 'unknown'}', 'retries': 1},`);
  lines.push('    start_date=datetime(2024, 1, 1),');
  lines.push('    schedule_interval=None,');
  lines.push(')');
  lines.push('');

  for (const task of spec.tasks) {
    lines.push(`${task.id} = KubernetesPodOperator(`);
    lines.push(`    task_id='${task.id}',`);
    lines.push(`    image='${task.image}',`);
    lines.push('    dag=dag,');
    lines.push(')');
    lines.push('');
  }

  for (const edge of spec.edges) {
    lines.push(`${edge.from} >> ${edge.to}`);
  }

  return lines.join('\n') + '\n';
}
