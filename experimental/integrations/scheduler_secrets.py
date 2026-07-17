"""
FlowForge Scheduler & Secrets Management Module

This module provides:
1. Environment-based secrets (for local development)
2. HashiCorp Vault integration stub (for production)
3. Cron-based pipeline scheduling
"""

import os
import json
from datetime import datetime
from typing import Optional, Dict, Any
import requests


class SecretsManager:
    """Manages secrets from environment variables or Vault."""
    
    def __init__(self, vault_enabled: bool = False, vault_addr: str = "http://localhost:8200", vault_token: str = None):
        self.vault_enabled = vault_enabled
        self.vault_addr = vault_addr
        self.vault_token = vault_token or os.getenv("VAULT_TOKEN")
    
    def get_secret(self, key: str, default: Optional[str] = None) -> Optional[str]:
        """Retrieve a secret by key."""
        # Try Vault first if enabled
        if self.vault_enabled and self.vault_token:
            try:
                headers = {"X-Vault-Token": self.vault_token}
                resp = requests.get(
                    f"{self.vault_addr}/v1/secret/data/{key}",
                    headers=headers,
                    timeout=5
                )
                if resp.status_code == 200:
                    return resp.json()["data"]["data"]["value"]
            except Exception as e:
                print(f"Vault fetch failed for {key}: {e}, falling back to env")
        
        # Fall back to environment variable
        env_key = key.upper().replace("-", "_").replace("/", "_")
        return os.getenv(env_key, default)
    
    def set_secret(self, key: str, value: str) -> bool:
        """Set a secret (Vault only in production)."""
        if self.vault_enabled and self.vault_token:
            try:
                headers = {"X-Vault-Token": self.vault_token}
                data = {"data": {"data": {"value": value}}}
                resp = requests.post(
                    f"{self.vault_addr}/v1/secret/data/{key}",
                    headers=headers,
                    json=data,
                    timeout=5
                )
                return resp.status_code == 200
            except Exception as e:
                print(f"Vault set failed for {key}: {e}")
                return False
        
        # Fall back to environment (not persistent)
        os.environ[key.upper().replace("-", "_")] = value
        return True


class PipelineScheduler:
    """Manages pipeline execution scheduling using cron expressions."""
    
    def __init__(self):
        self.schedules: Dict[str, Dict[str, Any]] = {}
    
    def add_schedule(self, pipeline_id: str, cron: str, enabled: bool = True) -> bool:
        """Add a cron schedule for a pipeline.
        
        Examples:
            "0 9 * * *"   - Every day at 9 AM
            "0 */6 * * *" - Every 6 hours
            "0 0 * * 0"   - Every Sunday at midnight
        """
        self.schedules[pipeline_id] = {
            "cron": cron,
            "enabled": enabled,
            "created_at": datetime.utcnow().isoformat(),
            "last_run": None,
            "next_run": None,
        }
        return True
    
    def get_schedule(self, pipeline_id: str) -> Optional[Dict[str, Any]]:
        """Get schedule for a pipeline."""
        return self.schedules.get(pipeline_id)
    
    def list_schedules(self) -> Dict[str, Dict[str, Any]]:
        """List all schedules."""
        return self.schedules
    
    def trigger_pipeline(self, pipeline_id: str, observability_url: str = "http://localhost:8000") -> bool:
        """Manually trigger a pipeline execution."""
        import uuid
        try:
            run_id = str(uuid.uuid4())
            requests.post(
                f"{observability_url}/runs",
                json={
                    "id": run_id,
                    "pipeline_id": pipeline_id,
                    "status": "triggered",
                    "started_at": datetime.utcnow().isoformat(),
                    "meta": "scheduled_trigger"
                },
                timeout=5
            )
            print(f"Pipeline {pipeline_id} triggered with run_id {run_id}")
            return True
        except Exception as e:
            print(f"Failed to trigger pipeline: {e}")
            return False


# Example usage
if __name__ == "__main__":
    # Initialize secrets manager
    secrets = SecretsManager(vault_enabled=False)
    
    # Get secrets from environment
    db_password = secrets.get_secret("db-password", default="default_pass")
    api_key = secrets.get_secret("api-key", default="test_key_123")
    print(f"DB Password: {db_password}")
    print(f"API Key: {api_key}")
    
    # Initialize scheduler
    scheduler = PipelineScheduler()
    
    # Add schedules
    scheduler.add_schedule("etl-pipeline", "0 9 * * *")  # Daily at 9 AM
    scheduler.add_schedule("analytics-pipeline", "0 */6 * * *")  # Every 6 hours
    
    # List schedules
    print("\nSchedules:")
    for pipeline_id, schedule in scheduler.list_schedules().items():
        print(f"  {pipeline_id}: {schedule['cron']} (enabled: {schedule['enabled']})")
    
    # Trigger a pipeline
    print("\nTriggering pipeline...")
    scheduler.trigger_pipeline("etl-pipeline")
