import ast
from dataclasses import dataclass
from typing import Dict, List, Optional


@dataclass
class PipelineMeta:
    name: str
    version: Optional[str] = None
    owner: Optional[str] = None


@dataclass
class TaskSpec:
    id: str
    function_name: str
    image: str
    timeout: Optional[str] = None
    retries: Optional[int] = None


@dataclass
class EdgeSpec:
    source: str
    target: str


@dataclass
class PipelineSpec:
    function_name: str
    meta: PipelineMeta
    tasks: List[TaskSpec]
    edges: List[EdgeSpec]


def _decorator_name(decorator: ast.AST) -> Optional[str]:
    if isinstance(decorator, ast.Name):
        return decorator.id
    if isinstance(decorator, ast.Call):
        if isinstance(decorator.func, ast.Name):
            return decorator.func.id
    return None


def _literal(value: ast.AST) -> Optional[object]:
    try:
        return ast.literal_eval(value)
    except Exception:
        return None


def _keyword_args(call: ast.Call) -> Dict[str, object]:
    out: Dict[str, object] = {}
    for kw in call.keywords:
        if kw.arg is None:
            continue
        lit = _literal(kw.value)
        if lit is not None:
            out[kw.arg] = lit
    return out


def _node_name(node: ast.AST) -> Optional[str]:
    if isinstance(node, ast.Name):
        return node.id
    if isinstance(node, ast.Constant) and isinstance(node.value, str):
        return node.value
    return None


def parse_pipeline(source: str) -> PipelineSpec:
    tree = ast.parse(source)

    pipeline_fn: Optional[str] = None
    pipeline_kwargs: Dict[str, object] = {}
    task_by_function: Dict[str, TaskSpec] = {}
    task_params: Dict[str, List[str]] = {}

    for node in tree.body:
        if not isinstance(node, ast.FunctionDef):
            continue
        task_params[node.name] = [arg.arg for arg in node.args.args]
        for deco in node.decorator_list:
            dname = _decorator_name(deco)
            if dname == 'pipeline' and pipeline_fn is None:
                pipeline_fn = node.name
                if isinstance(deco, ast.Call):
                    pipeline_kwargs = _keyword_args(deco)
            elif dname == 'task':
                kwargs = _keyword_args(deco) if isinstance(deco, ast.Call) else {}
                retries_raw = kwargs.get('retries')
                retries = retries_raw if isinstance(retries_raw, int) else None
                timeout_raw = kwargs.get('timeout')
                timeout = timeout_raw if isinstance(timeout_raw, str) else None
                image_raw = kwargs.get('image')
                image = image_raw if isinstance(image_raw, str) else 'python:3.11'
                task_by_function[node.name] = TaskSpec(
                    id=node.name,
                    function_name=node.name,
                    image=image,
                    timeout=timeout,
                    retries=retries,
                )

    if pipeline_fn is None:
        raise ValueError('Missing @pipeline decorator or pipeline function definition.')

    meta = PipelineMeta(
        name=str(pipeline_kwargs.get('name', pipeline_fn)),
        version=str(pipeline_kwargs['version']) if 'version' in pipeline_kwargs else None,
        owner=str(pipeline_kwargs.get('owner', 'unknown')),
    )

    pipeline_aliases = {pipeline_fn}
    for node in tree.body:
        if isinstance(node, ast.Assign) and len(node.targets) == 1 and isinstance(node.targets[0], ast.Name):
            val_name = _node_name(node.value)
            if val_name in pipeline_aliases:
                pipeline_aliases.add(node.targets[0].id)

    function_to_task_id: Dict[str, str] = {}
    raw_edges: List[EdgeSpec] = []

    for node in ast.walk(tree):
        if not isinstance(node, ast.Call):
            continue
        if not isinstance(node.func, ast.Attribute) or not isinstance(node.func.value, ast.Name):
            continue
        caller = node.func.value.id
        method = node.func.attr
        if caller not in pipeline_aliases:
            continue

        if method == 'add_task' and len(node.args) >= 2:
            task_id = _node_name(node.args[0])
            fn_ref = _node_name(node.args[1])
            if isinstance(task_id, str) and isinstance(fn_ref, str):
                function_to_task_id[fn_ref] = task_id

        if method == 'add_edge' and len(node.args) >= 3:
            src = _node_name(node.args[0])
            dst = _node_name(node.args[2])
            if isinstance(src, str) and isinstance(dst, str):
                raw_edges.append(EdgeSpec(source=src, target=dst))

    tasks: List[TaskSpec] = []
    task_id_by_function: Dict[str, str] = {}
    for fn_name, task in task_by_function.items():
        tid = function_to_task_id.get(fn_name, task.id)
        task_id_by_function[fn_name] = tid
        tasks.append(
            TaskSpec(
                id=tid,
                function_name=task.function_name,
                image=task.image,
                timeout=task.timeout,
                retries=task.retries,
            )
        )

    edges: List[EdgeSpec] = []
    seen = set()

    # explicit edges
    for edge in raw_edges:
        src = task_id_by_function.get(edge.source, edge.source)
        dst = task_id_by_function.get(edge.target, edge.target)
        key = (src, dst)
        if key not in seen:
            seen.add(key)
            edges.append(EdgeSpec(source=src, target=dst))

    # inferred edges from task parameters
    id_set = {t.id for t in tasks}
    fn_set = set(task_by_function.keys())
    for task in tasks:
        params = task_params.get(task.function_name, [])
        for p in params:
            src = None
            if p in id_set:
                src = p
            elif p in fn_set:
                src = task_id_by_function.get(p, p)
            if src and src != task.id:
                key = (src, task.id)
                if key not in seen:
                    seen.add(key)
                    edges.append(EdgeSpec(source=src, target=task.id))

    return PipelineSpec(function_name=pipeline_fn, meta=meta, tasks=tasks, edges=edges)
