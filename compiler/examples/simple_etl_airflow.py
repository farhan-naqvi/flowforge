"""FlowForge-generated Airflow DAG"""

from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.operators.bash import BashOperator
from airflow.providers.cncf.kubernetes.operators.kubernetes_pod import KubernetesPodOperator
from airflow.models import Variable

# DAG Configuration
dag_id = 'simple_etl'
pipeline_name = 'simple_etl'
pipeline_version = '1.0.0'
pipeline_owner = 'data_team'

default_args = {
    'owner': pipeline_owner,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
    'start_date': datetime(2024, 1, 1),
}

dag = DAG(
    dag_id='simple_etl',
    description='FlowForge pipeline: simple_etl',
    default_args=default_args,
    schedule_interval=None,
    catchup=False,
    tags=['flowforge', 'simple_etl'],
)

# Task Definitions
# Task: extract
extract = KubernetesPodOperator(
    task_id='extract',
    image='python:3.11',
    cmds=['python', 'extract.py'],
    namespace='default',
    dag=dag,
)

# Task: transform
transform = KubernetesPodOperator(
    task_id='transform',
    image='python:3.11',
    cmds=['python', 'transform.py'],
    namespace='default',
    dag=dag,
)

# Task: load
load = KubernetesPodOperator(
    task_id='load',
    image='python:3.11',
    cmds=['python', 'load.py'],
    namespace='default',
    dag=dag,
)

# Task Dependencies
extract >> transform
transform >> load
