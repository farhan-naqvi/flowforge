// ui/src/hooks/usePipelineEditor.ts - React hook for pipeline management

import { useState, useCallback, useReducer } from 'react';
import { Pipeline, Task, Edge, EditorMode, ValidationError, CompilationResult, ExecutionResult } from '../types/flowforge';
import { compilerService } from '../services/compilerService';

interface EditorState {
  pipeline: Pipeline | null;
  mode: EditorMode;
  selectedNode: string | null;
  selectedEdge: string | null;
  isDirty: boolean;
  validationErrors: ValidationError[];
  compilationResult: CompilationResult | null;
  executionResult: ExecutionResult | null;
  isLoading: boolean;
  error: string | null;
}

type EditorAction =
  | { type: 'SET_PIPELINE'; payload: Pipeline }
  | { type: 'ADD_TASK'; payload: { id: string; task: Task } }
  | { type: 'UPDATE_TASK'; payload: { id: string; task: Partial<Task> } }
  | { type: 'DELETE_TASK'; payload: string }
  | { type: 'ADD_EDGE'; payload: Edge }
  | { type: 'DELETE_EDGE'; payload: Edge }
  | { type: 'SELECT_NODE'; payload: string | null }
  | { type: 'SELECT_EDGE'; payload: string | null }
  | { type: 'SET_MODE'; payload: EditorMode }
  | { type: 'SET_VALIDATION'; payload: ValidationError[] }
  | { type: 'SET_COMPILATION'; payload: CompilationResult }
  | { type: 'SET_EXECUTION'; payload: ExecutionResult }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'MARK_DIRTY' }
  | { type: 'MARK_CLEAN' }
  | { type: 'RESET' };

function editorReducer(state: EditorState, action: EditorAction): EditorState {
  switch (action.type) {
    case 'SET_PIPELINE':
      return { ...state, pipeline: action.payload, isDirty: false };

    case 'ADD_TASK':
      if (!state.pipeline) return state;
      return {
        ...state,
        pipeline: {
          ...state.pipeline,
          tasks: {
            ...state.pipeline.tasks,
            [action.payload.id]: action.payload.task,
          },
        },
        isDirty: true,
      };

    case 'UPDATE_TASK':
      if (!state.pipeline) return state;
      const existingTask = state.pipeline.tasks[action.payload.id];
      return {
        ...state,
        pipeline: {
          ...state.pipeline,
          tasks: {
            ...state.pipeline.tasks,
            [action.payload.id]: { ...existingTask, ...action.payload.task },
          },
        },
        isDirty: true,
      };

    case 'DELETE_TASK': {
      if (!state.pipeline) return state;
      const { [action.payload]: _, ...remainingTasks } = state.pipeline.tasks;
      const filteredEdges = state.pipeline.edges.filter(
        e => e.from !== action.payload && e.to !== action.payload
      );
      return {
        ...state,
        pipeline: {
          ...state.pipeline,
          tasks: remainingTasks,
          edges: filteredEdges,
        },
        isDirty: true,
      };
    }

    case 'ADD_EDGE':
      if (!state.pipeline) return state;
      return {
        ...state,
        pipeline: {
          ...state.pipeline,
          edges: [...state.pipeline.edges, action.payload],
        },
        isDirty: true,
      };

    case 'DELETE_EDGE':
      if (!state.pipeline) return state;
      return {
        ...state,
        pipeline: {
          ...state.pipeline,
          edges: state.pipeline.edges.filter(
            e => !(e.from === action.payload.from && e.to === action.payload.to)
          ),
        },
        isDirty: true,
      };

    case 'SELECT_NODE':
      return { ...state, selectedNode: action.payload, selectedEdge: null };

    case 'SELECT_EDGE':
      return { ...state, selectedEdge: action.payload, selectedNode: null };

    case 'SET_MODE':
      return { ...state, mode: action.payload };

    case 'SET_VALIDATION':
      return { ...state, validationErrors: action.payload };

    case 'SET_COMPILATION':
      return { ...state, compilationResult: action.payload };

    case 'SET_EXECUTION':
      return { ...state, executionResult: action.payload };

    case 'SET_LOADING':
      return { ...state, isLoading: action.payload };

    case 'SET_ERROR':
      return { ...state, error: action.payload };

    case 'MARK_DIRTY':
      return { ...state, isDirty: true };

    case 'MARK_CLEAN':
      return { ...state, isDirty: false };

    case 'RESET':
      return {
        ...state,
        pipeline: null,
        selectedNode: null,
        selectedEdge: null,
        isDirty: false,
        validationErrors: [],
        compilationResult: null,
        executionResult: null,
      };

    default:
      return state;
  }
}

const initialState: EditorState = {
  pipeline: null,
  mode: 'dag',
  selectedNode: null,
  selectedEdge: null,
  isDirty: false,
  validationErrors: [],
  compilationResult: null,
  executionResult: null,
  isLoading: false,
  error: null,
};

export function usePipelineEditor() {
  const [state, dispatch] = useReducer(editorReducer, initialState);

  const setPipeline = useCallback((pipeline: Pipeline) => {
    dispatch({ type: 'SET_PIPELINE', payload: pipeline });
  }, []);

  const addTask = useCallback((id: string, task: Task) => {
    dispatch({ type: 'ADD_TASK', payload: { id, task } });
  }, []);

  const updateTask = useCallback((id: string, updates: Partial<Task>) => {
    dispatch({ type: 'UPDATE_TASK', payload: { id, task: updates } });
  }, []);

  const deleteTask = useCallback((id: string) => {
    dispatch({ type: 'DELETE_TASK', payload: id });
  }, []);

  const addEdge = useCallback((edge: Edge) => {
    dispatch({ type: 'ADD_EDGE', payload: edge });
  }, []);

  const deleteEdge = useCallback((edge: Edge) => {
    dispatch({ type: 'DELETE_EDGE', payload: edge });
  }, []);

  const selectNode = useCallback((nodeId: string | null) => {
    dispatch({ type: 'SELECT_NODE', payload: nodeId });
  }, []);

  const selectEdge = useCallback((edgeId: string | null) => {
    dispatch({ type: 'SELECT_EDGE', payload: edgeId });
  }, []);

  const setMode = useCallback((mode: EditorMode) => {
    dispatch({ type: 'SET_MODE', payload: mode });
  }, []);

  const validate = useCallback(async () => {
    if (!state.pipeline) return;
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const errors = await compilerService.validate(state.pipeline);
      dispatch({ type: 'SET_VALIDATION', payload: errors });
    } catch (err) {
      dispatch({ type: 'SET_ERROR', payload: err instanceof Error ? err.message : 'Validation failed' });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, [state.pipeline]);

  const compile = useCallback(async (executor: 'argo' | 'airflow') => {
    if (!state.pipeline) return;
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const result = await compilerService.compile(state.pipeline, executor);
      dispatch({ type: 'SET_COMPILATION', payload: result });
    } catch (err) {
      dispatch({ type: 'SET_ERROR', payload: err instanceof Error ? err.message : 'Compilation failed' });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, [state.pipeline]);

  const exportPipeline = useCallback((format: 'json' | 'yaml' | 'py') => {
    if (!state.pipeline) return '';
    return compilerService.exportPipeline(state.pipeline, format);
  }, [state.pipeline]);

  const importPipeline = useCallback(async (content: string, format: 'json' | 'yaml' | 'py') => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const pipeline = await compilerService.importPipeline(content, format);
      dispatch({ type: 'SET_PIPELINE', payload: pipeline });
    } catch (err) {
      dispatch({ type: 'SET_ERROR', payload: err instanceof Error ? err.message : 'Import failed' });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  const reset = useCallback(() => {
    dispatch({ type: 'RESET' });
  }, []);

  return {
    // State
    pipeline: state.pipeline,
    mode: state.mode,
    selectedNode: state.selectedNode,
    selectedEdge: state.selectedEdge,
    isDirty: state.isDirty,
    validationErrors: state.validationErrors,
    compilationResult: state.compilationResult,
    executionResult: state.executionResult,
    isLoading: state.isLoading,
    error: state.error,

    // Actions
    setPipeline,
    addTask,
    updateTask,
    deleteTask,
    addEdge,
    deleteEdge,
    selectNode,
    selectEdge,
    setMode,
    validate,
    compile,
    exportPipeline,
    importPipeline,
    reset,
  };
}
