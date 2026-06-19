from ..parser import PipelineSpec


def compile_airflow(spec: PipelineSpec) -> str:
    lines = [
        'from airflow import DAG',
        'from airflow.providers.cncf.kubernetes.operators.kubernetes_pod import KubernetesPodOperator',
        'from datetime import datetime, timedelta',
        '',
        'dag = DAG(',
        f"    dag_id='{spec.meta.name}',",
        f"    default_args={{'owner': '{spec.meta.owner or 'unknown'}', 'retries': 1}},",
        '    start_date=datetime(2024, 1, 1),',
        '    schedule_interval=None,',
        ')',
        '',
    ]

    for task in spec.tasks:
        lines.extend([
            f'{task.id} = KubernetesPodOperator(',
            f"    task_id='{task.id}',",
            f"    image='{task.image}',",
            '    dag=dag,',
            ')',
            '',
        ])

    for edge in spec.edges:
        lines.append(f'{edge.source} >> {edge.target}')

    return '\n'.join(lines) + '\n'
