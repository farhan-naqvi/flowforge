"""FlowForge-generated Airflow DAG - Fan Out / Fan In Pattern"""

from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.providers.cncf.kubernetes.operators.kubernetes_pod import KubernetesPodOperator

# DAG Configuration
dag_id = 'fan_out_fan_in'
pipeline_name = 'fan_out_fan_in'
pipeline_version = '1.0.0'
pipeline_owner = 'data_team'

default_args = {
    'owner': pipeline_owner,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
    'start_date': datetime(2024, 1, 1),
}

dag = DAG(
    dag_id='fan_out_fan_in',
    description='FlowForge pipeline: fan_out_fan_in',
    default_args=default_args,
    schedule_interval=None,
    catchup=False,
    tags=['flowforge', 'fan_out_fan_in'],
)

# Task Definitions
# Task: source (entry point)
source = KubernetesPodOperator(
    task_id='source',
    image='python:3.11',
    cmds=['python', 'source.py'],
    namespace='default',
    dag=dag,
)

# Task: process_a (parallel)
process_a = KubernetesPodOperator(
    task_id='process_a',
    image='python:3.11',
    cmds=['python', 'process_a.py'],
    namespace='default',
    dag=dag,
)

# Task: process_b (parallel)
process_b = KubernetesPodOperator(
    task_id='process_b',
    image='python:3.11',
    cmds=['python', 'process_b.py'],
    namespace='default',
    dag=dag,
)

# Task: merge (waits for both branches)
merge = KubernetesPodOperator(
    task_id='merge',
    image='python:3.11',
    cmds=['python', 'merge.py'],
    namespace='default',
    dag=dag,
)

# Task Dependencies - Fan out from source
source >> [process_a, process_b]

# Fan in to merge
[process_a, process_b] >> merge
