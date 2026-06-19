// ui/src/types/flowforge.ts - TypeScript types for FlowForge UI

export interface Pipeline {
  id: string;
  name: string;
  version: string;
  description: string;
  tasks: Record<string, Task>;
  edges: Edge[];
  metadata: Record<string, any>;
  createdAt: Date;
  updatedAt: Date;
}

export interface Task {
  id: string;
  name: string;
  handler: Handler;
  config: Config;
  inputs: Input[];
  outputs: Output[];
  dependencies: string[];
}

export interface Handler {
  type: string; // 'python', 'bash', 'spark', etc.
  command: string;
  script?: string;
}

export interface Config {
  image: string;
  resources?: {
    memory: string;
    cpu: string;
    gpu?: string;
  };
  timeout?: number;
  retries?: number;
  environment?: Record<string, string>;
}

export interface Input {
  name: string;
  type: string;
  required: boolean;
  default?: any;
}

export interface Output {
  name: string;
  type: string;
  description?: string;
}

export interface Edge {
  from: string;
  to: string;
  fromPort?: string;
  toPort?: string;
  condition?: string;
}

export interface DAGNode {
  id: string;
  label: string;
  type: string;
  position: { x: number; y: number };
  data: any;
  inputs: Port[];
  outputs: Port[];
}

export interface Port {
  id: string;
  name: string;
  type: string;
}

export interface DAGEdge {
  id: string;
  source: string;
  target: string;
  sourceHandle: string;
  targetHandle: string;
}

export interface CompilationResult {
  executor: string; // 'argo' | 'airflow'
  yaml?: string;
  python?: string;
  errors: string[];
  warnings: string[];
}

export interface ExecutionResult {
  id: string;
  pipelineId: string;
  status: string;
  startedAt: Date;
  completedAt?: Date;
  duration?: number;
  tasks: Record<string, TaskExecutionResult>;
}

export interface TaskExecutionResult {
  taskId: string;
  status: string;
  exitCode?: number;
  logs?: string;
  duration?: number;
}

export interface ValidationError {
  path: string[];
  message: string;
  code: string;
}

export type EditorMode = 'dag' | 'yaml' | 'sdk';

export interface UIState {
  pipeline: Pipeline | null;
  mode: EditorMode;
  selectedNode: string | null;
  selectedEdge: string | null;
  compilationResult: CompilationResult | null;
  executionResult: ExecutionResult | null;
  validationErrors: ValidationError[];
  isDirty: boolean;
}
