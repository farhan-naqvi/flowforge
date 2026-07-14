export interface PipelineMeta {
  name: string;
  version?: string;
  owner?: string;
}

export interface TaskSpec {
  id: string;
  functionName: string;
  image: string;
  timeout?: string;
  retries?: number;
}

export interface EdgeSpec {
  from: string;
  to: string;
}

export interface PipelineSpec {
  pipelineFunctionName: string;
  meta: PipelineMeta;
  tasks: TaskSpec[];
  edges: EdgeSpec[];
}

function parseDecoratorArgs(input: string): Record<string, string> {
  const out: Record<string, string> = {};
  const re = /(\w+)\s*=\s*("[^"]*"|'[^']*'|[^,\)\s]+)/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(input)) !== null) {
    const key = m[1];
    let value = m[2].trim();
    if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) {
      value = value.slice(1, -1);
    }
    out[key] = value;
  }
  return out;
}

export function parsePythonPipeline(source: string): PipelineSpec {
  const pipelineMatch = /@pipeline(?:\(([^)]*)\))?\s*def\s+(\w+)\s*\(/s.exec(source);
  if (!pipelineMatch) {
    throw new Error('Missing @pipeline decorator or pipeline function definition.');
  }

  const pipelineArgs = parseDecoratorArgs(pipelineMatch[1] || '');
  const pipelineFunctionName = pipelineMatch[2];
  const meta: PipelineMeta = {
    name: pipelineArgs.name || pipelineFunctionName,
    version: pipelineArgs.version,
    owner: pipelineArgs.owner || 'unknown'
  };

  const taskByFunction = new Map<string, Omit<TaskSpec, 'id'>>();
  const taskParams = new Map<string, string[]>();
  const taskDefRegex = /@task(?:\(([^)]*)\))?\s*def\s+(\w+)\s*\(([^)]*)\)\s*(?:->\s*[^:]+)?\s*:/g;
  let td: RegExpExecArray | null;
  while ((td = taskDefRegex.exec(source)) !== null) {
    const args = parseDecoratorArgs(td[1] || '');
    const fn = td[2];
    const paramsRaw = td[3] || '';
    const params = paramsRaw
      .split(',')
      .map(p => p.trim())
      .filter(Boolean)
      .map(p => p.split(':')[0].trim())
      .map(p => p.replace(/^\*+/, ''));
    taskParams.set(fn, params);
    taskByFunction.set(fn, {
      functionName: fn,
      image: args.image || 'python:3.11',
      timeout: args.timeout,
      retries: args.retries ? Number(args.retries) : undefined
    });
  }

  const pipelineAliases = new Set<string>([pipelineFunctionName]);
  const aliasRegex = /(\w+)\s*=\s*(\w+)/g;
  let am: RegExpExecArray | null;
  while ((am = aliasRegex.exec(source)) !== null) {
    const left = am[1];
    const right = am[2];
    if (pipelineAliases.has(right)) {
      pipelineAliases.add(left);
    }
  }

  const functionToTaskId = new Map<string, string>();
  const addTaskRegex = /(\w+)\.add_task\(\s*["']([^"']+)["']\s*,\s*(\w+)\s*\)/g;
  let at: RegExpExecArray | null;
  while ((at = addTaskRegex.exec(source)) !== null) {
    const pipelineVar = at[1];
    if (!pipelineAliases.has(pipelineVar)) {
      continue;
    }
    functionToTaskId.set(at[3], at[2]);
  }

  const tasks: TaskSpec[] = [];
  for (const [fn, base] of taskByFunction.entries()) {
    const taskId = functionToTaskId.get(fn) || fn;
    tasks.push({
      id: taskId,
      functionName: base.functionName,
      image: base.image,
      timeout: base.timeout,
      retries: base.retries
    });
  }

  const edges: EdgeSpec[] = [];
  const seen = new Set<string>();
  const edgeRegex = /(\w+)\.add_edge\(\s*([\w"']+)\s*,\s*["'][^"']*["']\s*,\s*([\w"']+)\s*,\s*["'][^"']*["']\s*\)/g;
  let em: RegExpExecArray | null;
  while ((em = edgeRegex.exec(source)) !== null) {
    const pipelineVar = em[1];
    if (!pipelineAliases.has(pipelineVar)) {
      continue;
    }
    const fromRaw = em[2].replace(/^['"]|['"]$/g, '');
    const toRaw = em[3].replace(/^['"]|['"]$/g, '');
    const from = functionToTaskId.get(fromRaw) || fromRaw;
    const to = functionToTaskId.get(toRaw) || toRaw;
    const key = `${from}->${to}`;
    if (!seen.has(key)) {
      seen.add(key);
      edges.push({ from, to });
    }
  }

  // infer edges from task parameter names when explicit edges are absent
  const taskIds = new Set(tasks.map(t => t.id));
  const taskFnNames = new Set(tasks.map(t => t.functionName));
  for (const t of tasks) {
    const params = taskParams.get(t.functionName) || [];
    for (const p of params) {
      let src: string | undefined;
      if (taskIds.has(p)) {
        src = p;
      } else if (taskFnNames.has(p)) {
        src = functionToTaskId.get(p) || p;
      }
      if (src && src !== t.id) {
        const key = `${src}->${t.id}`;
        if (!seen.has(key)) {
          seen.add(key);
          edges.push({ from: src, to: t.id });
        }
      }
    }
  }

  return { pipelineFunctionName, meta, tasks, edges };
}
