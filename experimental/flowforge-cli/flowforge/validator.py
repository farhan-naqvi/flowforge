from typing import Dict, List

from .parser import PipelineSpec


def validate_pipeline(spec: PipelineSpec) -> List[str]:
    errors: List[str] = []
    task_ids = {task.id for task in spec.tasks}

    if len(task_ids) != len(spec.tasks):
        errors.append('Duplicate task IDs are not allowed.')

    if not spec.meta.name:
        errors.append('Pipeline name is required.')

    if not task_ids:
        errors.append('Pipeline must define at least one @task.')

    for edge in spec.edges:
        if edge.source not in task_ids:
            errors.append(f'Edge source task not found: {edge.source}')
        if edge.target not in task_ids:
            errors.append(f'Edge destination task not found: {edge.target}')

    # Cycle detection
    adj: Dict[str, List[str]] = {}
    for edge in spec.edges:
        adj.setdefault(edge.source, []).append(edge.target)

    color: Dict[str, int] = {k: 0 for k in task_ids}

    def visit(node: str) -> bool:
        color[node] = 1
        for nxt in adj.get(node, []):
            c = color.get(nxt, 0)
            if c == 1:
                return True
            if c == 0 and visit(nxt):
                return True
        color[node] = 2
        return False

    for node in task_ids:
        if color.get(node, 0) == 0 and visit(node):
            errors.append('Pipeline contains a cycle.')
            break

    return errors
