from pathlib import Path
from collections import defaultdict, deque
from typing import Optional

import click

from .compilers.airflow import compile_airflow
from .compilers.argo import compile_argo
from .parser import parse_pipeline
from .validator import validate_pipeline


def _read_file(path: Path) -> str:
    return path.read_text(encoding='utf-8')


def _write_text(path: Path, content: str) -> None:
    with path.open('w', encoding='utf-8', newline='\n') as f:
        f.write(content)


def _default_output(input_path: Path, target: str) -> Path:
    suffix = '.argo.yaml' if target == 'argo' else '.airflow.py'
    return input_path.with_name(f"{input_path.stem}{suffix}")


@click.group()
def main() -> None:
    """FlowForge pipeline compiler CLI."""


@main.group()
def generate() -> None:
    """Generate deployable outputs."""


@generate.command('argo')
@click.argument('pipeline_file', type=click.Path(exists=True, path_type=Path))
@click.option('--output', '-o', type=click.Path(path_type=Path), default=None)
def generate_argo(pipeline_file: Path, output: Optional[Path]) -> None:
    spec = parse_pipeline(_read_file(pipeline_file))
    errors = validate_pipeline(spec)
    if errors:
        for err in errors:
            click.echo(f'ERROR: {err}')
        raise SystemExit(1)
    rendered = compile_argo(spec)
    out = output or _default_output(pipeline_file, 'argo')
    _write_text(out, rendered)
    click.echo(f'Wrote {out}')


@generate.command('airflow')
@click.argument('pipeline_file', type=click.Path(exists=True, path_type=Path))
@click.option('--output', '-o', type=click.Path(path_type=Path), default=None)
def generate_airflow(pipeline_file: Path, output: Optional[Path]) -> None:
    spec = parse_pipeline(_read_file(pipeline_file))
    errors = validate_pipeline(spec)
    if errors:
        for err in errors:
            click.echo(f'ERROR: {err}')
        raise SystemExit(1)
    rendered = compile_airflow(spec)
    out = output or _default_output(pipeline_file, 'airflow')
    _write_text(out, rendered)
    click.echo(f'Wrote {out}')


@main.command('validate')
@click.argument('pipeline_file', type=click.Path(exists=True, path_type=Path))
def validate_cmd(pipeline_file: Path) -> None:
    spec = parse_pipeline(_read_file(pipeline_file))
    errors = validate_pipeline(spec)
    if errors:
        for err in errors:
            click.echo(f'ERROR: {err}')
        raise SystemExit(1)
    click.echo('Pipeline is valid.')


@main.command('preview')
@click.argument('pipeline_file', type=click.Path(exists=True, path_type=Path))
def preview_cmd(pipeline_file: Path) -> None:
    spec = parse_pipeline(_read_file(pipeline_file))
    errors = validate_pipeline(spec)
    if errors:
        for err in errors:
            click.echo(f'ERROR: {err}')
        raise SystemExit(1)

    click.echo(f'Pipeline: {spec.meta.name} ({spec.meta.version or "n/a"})')
    click.echo(f'Owner: {spec.meta.owner or "unknown"}')
    click.echo('')

    indegree = defaultdict(int)
    adj = defaultdict(list)
    for task in spec.tasks:
        indegree[task.id] += 0
    for edge in spec.edges:
        adj[edge.source].append(edge.target)
        indegree[edge.target] += 1

    q = deque([t.id for t in spec.tasks if indegree[t.id] == 0])
    order = []
    while q:
        n = q.popleft()
        order.append(n)
        for nxt in adj[n]:
            indegree[nxt] -= 1
            if indegree[nxt] == 0:
                q.append(nxt)

    for idx, task_id in enumerate(order):
        click.echo(task_id)
        if idx < len(order) - 1:
            click.echo('   ↓')


if __name__ == '__main__':
    main()
