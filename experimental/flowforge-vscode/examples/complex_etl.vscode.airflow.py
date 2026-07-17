from airflow import DAG
from airflow.providers.cncf.kubernetes.operators.kubernetes_pod import KubernetesPodOperator
from datetime import datetime, timedelta

dag = DAG(
    dag_id='complex_etl',
    default_args={'owner': 'ops_team', 'retries': 1},
    start_date=datetime(2024, 1, 1),
    schedule_interval=None,
)

extract = KubernetesPodOperator(
    task_id='extract',
    image='python:3.11',
    dag=dag,
)

transform = KubernetesPodOperator(
    task_id='transform',
    image='python:3.11',
    dag=dag,
)

load = KubernetesPodOperator(
    task_id='load',
    image='python:3.11',
    dag=dag,
)

extract >> transform
transform >> load
