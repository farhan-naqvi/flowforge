from ..parser import PipelineSpec


def _timeout_seconds(timeout: str):
    if not timeout:
        return None
    timeout = timeout.strip()
    if timeout.endswith('s') and timeout[:-1].isdigit():
        return int(timeout[:-1])
    return None


def compile_argo(spec: PipelineSpec) -> str:
    deps_by_target = {}
    for edge in spec.edges:
        deps_by_target.setdefault(edge.target, []).append(edge.source)

    lines = [
        'apiVersion: argoproj.io/v1alpha1',
        'kind: Workflow',
        'metadata:',
        f'  name: {spec.meta.name}',
        '  namespace: default',
        'spec:',
    ]

    first_timeout = None
    for task in spec.tasks:
        first_timeout = _timeout_seconds(task.timeout)
        if first_timeout:
            break

    if first_timeout:
        lines.append(f'  activeDeadlineSeconds: {first_timeout}')

    lines.extend([
        f'  entrypoint: {spec.meta.name}',
        '  templates:',
        f'  - name: {spec.meta.name}',
        '    dag:',
        '      tasks:',
    ])

    for task in spec.tasks:
        lines.append(f'      - name: {task.id}')
        lines.append(f'        template: {task.id}')
        deps = deps_by_target.get(task.id, [])
        if deps:
            lines.append(f"        dependencies: [{', '.join(deps)}]")

    for task in spec.tasks:
        lines.extend([
            f'  - name: {task.id}',
            '    container:',
            f'      image: {task.image}',
            f'      command: ["python", "{task.id}.py"]',
            '      resources:',
            '        requests:',
            '          memory: "512Mi"',
            '          cpu: "100m"',
        ])
        if task.retries is not None:
            lines.extend([
                '    retryStrategy:',
                f'      limit: {task.retries}',
            ])

    return '\n'.join(lines) + '\n'
