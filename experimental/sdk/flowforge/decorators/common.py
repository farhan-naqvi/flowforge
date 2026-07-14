"""Built-in task decorators for common data operations."""

from __future__ import annotations
from typing import Any, Callable, Dict, List, Optional
from flowforge.core.task import task


# Source task decorators (data ingestion)

@task(image="python:3.11")
def kafka(topic: str, brokers: str = "localhost:9092") -> list:
    """Read data from Kafka topic.
    
    Args:
        topic: Kafka topic name
        brokers: Broker addresses
    
    Returns:
        List of records from Kafka
    """
    pass


@task(image="python:3.11")
def s3_read(bucket: str, key: str) -> dict:
    """Read data from S3.
    
    Args:
        bucket: S3 bucket name
        key: S3 object key
    
    Returns:
        Data from S3 object
    """
    pass


@task(image="python:3.11")
def sql_read(query: str, database: str = "default") -> list:
    """Read data from SQL database.
    
    Args:
        query: SQL query
        database: Database name
    
    Returns:
        Query results as list of records
    """
    pass


# Transform task decorators (data processing)

@task(image="python:3.11")
def transform(data: Any, image: Optional[str] = None) -> Any:
    """Generic transform task.
    
    Args:
        data: Input data
        image: Docker image override
    
    Returns:
        Transformed data
    """
    return data


@task(image="python:3.11")
def aggregate(data: list, key: Optional[str] = None) -> dict:
    """Aggregate data by key.
    
    Args:
        data: List of records
        key: Field to aggregate by
    
    Returns:
        Aggregated results
    """
    pass


@task(image="python:3.11")
def filter_data(data: list, condition: Optional[str] = None) -> list:
    """Filter data based on condition.
    
    Args:
        data: List of records
        condition: Filter condition
    
    Returns:
        Filtered records
    """
    return [record for record in data if record]


# Sink task decorators (data export)

@task(image="python:3.11")
def save(data: Any, path: str = "./output") -> None:
    """Save data to file.
    
    Args:
        data: Data to save
        path: Output file path
    """
    pass


@task(image="python:3.11")
def s3_write(data: Any, bucket: str, key: str) -> None:
    """Write data to S3.
    
    Args:
        data: Data to write
        bucket: S3 bucket name
        key: S3 object key
    """
    pass


@task(image="python:3.11")
def sql_write(data: list, table: str, database: str = "default") -> None:
    """Write data to SQL database.
    
    Args:
        data: List of records to write
        table: Table name
        database: Database name
    """
    pass


@task(image="python:3.11")
def notify(message: str, channel: str = "slack") -> None:
    """Send notification.
    
    Args:
        message: Notification message
        channel: Notification channel (slack, email, etc)
    """
    pass


__all__ = [
    'kafka', 's3_read', 'sql_read',
    'transform', 'aggregate', 'filter_data',
    'save', 's3_write', 'sql_write', 'notify',
]
