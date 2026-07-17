"""Per-target capability & semantic-loss analysis.

This is FlowForge's differentiating feature: before you commit a pipeline to an
orchestrator, it tells you exactly which IR features that target preserves,
approximates (lossy), or cannot represent (unsupported). Nothing is dropped
silently — every gap becomes a line in the capability report.

The support matrix is data, so adding a target later is a table entry plus a
compiler, not a rewrite.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Dict, List

from .ir import PipelineSpec

SUPPORTED = "supported"
LOSSY = "lossy"
UNSUPPORTED = "unsupported"

# feature -> level -> human note, per target.
CAPABILITIES: Dict[str, Dict[str, Dict[str, str]]] = {
    "argo": {
        "dag": {"level": SUPPORTED, "note": "task dependencies map to Argo DAG task deps"},
        "retry": {"level": SUPPORTED, "note": "retry.maxAttempts -> retryStrategy.limit"},
        "retry_backoff": {"level": SUPPORTED, "note": "backoff/multiplier -> retryStrategy.backoff"},
        "timeout": {"level": SUPPORTED, "note": "timeout -> template activeDeadlineSeconds"},
        "resources": {"level": SUPPORTED, "note": "resources -> container resources.requests"},
        "env": {"level": SUPPORTED, "note": "handler.env -> container env"},
        "python_handler": {"level": LOSSY, "note": "python source runs as a container command; needs an image"},
        "schedule": {"level": UNSUPPORTED, "note": "cron schedules require CronWorkflow (not emitted in v1)"},
        "conditional": {"level": UNSUPPORTED, "note": "conditional branching (when:) not emitted in v1"},
    },
    "airflow": {
        "dag": {"level": SUPPORTED, "note": "task dependencies map to >> chaining"},
        "retry": {"level": SUPPORTED, "note": "retry.maxAttempts -> task retries"},
        "retry_backoff": {"level": LOSSY, "note": "only retry_delay is set; exponential/multiplier approximated"},
        "timeout": {"level": SUPPORTED, "note": "timeout -> execution_timeout"},
        "resources": {"level": LOSSY, "note": "cpu/memory recorded as doc only unless KubernetesPodOperator"},
        "env": {"level": SUPPORTED, "note": "handler.env -> operator op_kwargs env"},
        "python_handler": {"level": SUPPORTED, "note": "python source runs in a PythonOperator callable"},
        "schedule": {"level": SUPPORTED, "note": "cron schedule -> DAG schedule="},
        "conditional": {"level": UNSUPPORTED, "note": "branching requires BranchPythonOperator (not emitted in v1)"},
    },
}

TARGETS = tuple(CAPABILITIES)


@dataclass
class CapabilityItem:
    feature: str
    level: str
    note: str
    used: bool  # is this feature actually present in the pipeline?

    def to_dict(self) -> dict:
        return {"feature": self.feature, "level": self.level, "note": self.note, "used": self.used}


@dataclass
class CapabilityReport:
    target: str
    items: List[CapabilityItem]

    @property
    def lossy_used(self) -> List[CapabilityItem]:
        return [i for i in self.items if i.used and i.level == LOSSY]

    @property
    def unsupported_used(self) -> List[CapabilityItem]:
        return [i for i in self.items if i.used and i.level == UNSUPPORTED]

    def to_dict(self) -> dict:
        return {
            "target": self.target,
            "items": [i.to_dict() for i in self.items],
            "summary": {
                "lossy": len(self.lossy_used),
                "unsupported": len(self.unsupported_used),
            },
        }


def _features_used(spec: PipelineSpec) -> Dict[str, bool]:
    used = {
        "dag": bool(spec.edges),
        "retry": False,
        "retry_backoff": False,
        "timeout": False,
        "resources": False,
        "env": False,
        "python_handler": False,
        "schedule": bool(spec.metadata.tags.get("schedule")),
        "conditional": any(t.type == "Conditional" for t in spec.tasks.values()),
    }
    for task in spec.tasks.values():
        if task.retry:
            used["retry"] = True
            if task.retry.backoff and task.retry.backoff != "fixed":
                used["retry_backoff"] = True
        if task.timeout:
            used["timeout"] = True
        if task.resources:
            used["resources"] = True
        if task.handler.env:
            used["env"] = True
        if task.handler.type == "python":
            used["python_handler"] = True
        if task.type == "Schedule":
            used["schedule"] = True
    return used


def analyze(spec: PipelineSpec, target: str) -> CapabilityReport:
    if target not in CAPABILITIES:
        raise ValueError(f"unknown target {target!r}; known: {', '.join(TARGETS)}")
    used = _features_used(spec)
    items = [
        CapabilityItem(feature=feat, level=info["level"], note=info["note"], used=used.get(feat, False))
        for feat, info in CAPABILITIES[target].items()
    ]
    return CapabilityReport(target=target, items=items)


def analyze_all(spec: PipelineSpec) -> Dict[str, CapabilityReport]:
    return {t: analyze(spec, t) for t in TARGETS}
