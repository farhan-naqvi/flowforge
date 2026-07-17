// ui/src/components/EditorModes.tsx - React components for DAG, YAML, and SDK editor modes

import React, { useState } from 'react';
import { Pipeline, Task, Edge } from '../types/flowforge';
import { usePipelineEditor } from '../hooks/usePipelineEditor';

// DAG Editor Component
export const DAGEditor: React.FC<{ pipeline: Pipeline | null; onSelect?: (nodeId: string) => void }> = ({
  pipeline,
  onSelect,
}) => {
  const [positions, setPositions] = useState<Record<string, { x: number; y: number }>>({});

  const getTaskPosition = (taskId: string, index: number) => {
    if (positions[taskId]) return positions[taskId];
    return { x: index * 200, y: 100 };
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', border: '1px solid #ccc' }}>
      <div style={{ padding: '10px', backgroundColor: '#f5f5f5', borderBottom: '1px solid #ccc' }}>
        <h3>Visual DAG Editor</h3>
        <p>Drag nodes to rearrange. Connect ports to create edges.</p>
      </div>

      <div style={{ flex: 1, position: 'relative', backgroundColor: '#fafafa', overflow: 'auto' }}>
        <svg style={{ width: '100%', height: '100%', position: 'absolute', top: 0, left: 0 }}>
          {/* Render edges (connections) */}
          {pipeline?.edges.map((edge, i) => {
            const fromPos = getTaskPosition(edge.from, 0);
            const toPos = getTaskPosition(edge.to, 1);
            return (
              <line
                key={`edge-${i}`}
                x1={fromPos.x + 100}
                y1={fromPos.y + 40}
                x2={toPos.x}
                y2={toPos.y}
                stroke="#666"
                strokeWidth="2"
                markerEnd="url(#arrowhead)"
              />
            );
          })}
          <defs>
            <marker id="arrowhead" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
              <polygon points="0 0, 10 3, 0 6" fill="#666" />
            </marker>
          </defs>
        </svg>

        {/* Render task nodes */}
        {pipeline &&
          Object.entries(pipeline.tasks).map(([taskId, task], index) => {
            const pos = getTaskPosition(taskId, index);
            return (
              <div
                key={taskId}
                onClick={() => onSelect?.(taskId)}
                style={{
                  position: 'absolute',
                  left: `${pos.x}px`,
                  top: `${pos.y}px`,
                  width: '150px',
                  padding: '10px',
                  backgroundColor: '#fff',
                  border: '2px solid #007bff',
                  borderRadius: '4px',
                  cursor: 'pointer',
                  userSelect: 'none',
                  boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
                }}
              >
                <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>{task.name}</div>
                <div style={{ fontSize: '12px', color: '#666' }}>{task.handler.type}</div>
                <div style={{ fontSize: '11px', color: '#999', marginTop: '4px' }}>{task.config.image}</div>
              </div>
            );
          })}
      </div>

      <div style={{ padding: '10px', backgroundColor: '#f5f5f5', borderTop: '1px solid #ccc', fontSize: '12px' }}>
        Tasks: {pipeline?.tasks ? Object.keys(pipeline.tasks).length : 0} | Edges:{' '}
        {pipeline?.edges ? pipeline.edges.length : 0}
      </div>
    </div>
  );
};

// YAML Editor Component
export const YAMLEditor: React.FC<{ initialYAML?: string; onChange?: (yaml: string) => void }> = ({
  initialYAML = '',
  onChange,
}) => {
  const [yaml, setYAML] = useState(initialYAML);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newYAML = e.target.value;
    setYAML(newYAML);
    onChange?.(newYAML);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ padding: '10px', backgroundColor: '#f5f5f5', borderBottom: '1px solid #ccc' }}>
        <h3>YAML Editor</h3>
        <p>Edit pipeline definition in YAML format</p>
      </div>

      <textarea
        value={yaml}
        onChange={handleChange}
        style={{
          flex: 1,
          fontFamily: 'Monaco, Courier New, monospace',
          fontSize: '13px',
          padding: '10px',
          border: 'none',
          resize: 'none',
          backgroundColor: '#f8f8f8',
        }}
        placeholder="metadata:
  name: my_pipeline
  version: 1.0.0

tasks:
  task1:
    handler:
      type: python
      command: python /scripts/transform.py
    config:
      image: python:3.11
      memory: 512M
      cpu: 1

edges:
  - from: task1
    to: task2"
      />

      <div style={{ padding: '10px', backgroundColor: '#f5f5f5', borderTop: '1px solid #ccc', fontSize: '12px' }}>
        Lines: {yaml.split('\n').length}
      </div>
    </div>
  );
};

// SDK Generator Component
export const SDKEditor: React.FC<{ initialPython?: string; onChange?: (python: string) => void }> = ({
  initialPython = '',
  onChange,
}) => {
  const [python, setPython] = useState(initialPython);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newPython = e.target.value;
    setPython(newPython);
    onChange?.(newPython);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ padding: '10px', backgroundColor: '#f5f5f5', borderBottom: '1px solid #ccc' }}>
        <h3>SDK Editor (Python)</h3>
        <p>Define pipeline using FlowForge Python SDK</p>
      </div>

      <textarea
        value={python}
        onChange={handleChange}
        style={{
          flex: 1,
          fontFamily: 'Monaco, Courier New, monospace',
          fontSize: '13px',
          padding: '10px',
          border: 'none',
          resize: 'none',
          backgroundColor: '#f8f8f8',
        }}
        placeholder="from flowforge import pipeline, task

@pipeline(name='my_pipeline', version='1.0.0')
def my_pipeline():
    @task(image='python:3.11', memory='512M')
    def transform():
        pass
    
    @task(image='python:3.11')
    def load():
        pass
    
    transform() >> load()

if __name__ == '__main__':
    my_pipeline()"
      />

      <div style={{ padding: '10px', backgroundColor: '#f5f5f5', borderTop: '1px solid #ccc', fontSize: '12px' }}>
        Lines: {python.split('\n').length}
      </div>
    </div>
  );
};

// Main Editor Component
export const PipelineEditor: React.FC<{ onCompile?: (yaml: string) => void }> = ({ onCompile }) => {
  const editor = usePipelineEditor();
  const [yamlContent, setYamlContent] = useState('');
  const [pythonContent, setPythonContent] = useState('');

  const handleModeChange = (mode: 'dag' | 'yaml' | 'sdk') => {
    if (mode === 'yaml' && editor.pipeline) {
      setYamlContent(editor.exportPipeline('yaml'));
    } else if (mode === 'sdk' && editor.pipeline) {
      setPythonContent(editor.exportPipeline('py'));
    }
    editor.setMode(mode);
  };

  const handleCompile = async (executor: 'argo' | 'airflow') => {
    await editor.compile(executor);
    if (editor.compilationResult) {
      const output = executor === 'argo' ? editor.compilationResult.yaml : editor.compilationResult.python;
      if (output) onCompile?.(output);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh', backgroundColor: '#fff' }}>
      {/* Header */}
      <div
        style={{
          padding: '15px 20px',
          backgroundColor: '#fff',
          borderBottom: '1px solid #e0e0e0',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <h1 style={{ margin: 0, fontSize: '24px', fontWeight: '600' }}>FlowForge Pipeline Editor</h1>
        <div style={{ display: 'flex', gap: '10px' }}>
          <button
            onClick={() => handleModeChange('dag')}
            style={{
              padding: '8px 16px',
              backgroundColor: editor.mode === 'dag' ? '#007bff' : '#f0f0f0',
              color: editor.mode === 'dag' ? '#fff' : '#000',
              border: '1px solid #ddd',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Visual DAG
          </button>
          <button
            onClick={() => handleModeChange('yaml')}
            style={{
              padding: '8px 16px',
              backgroundColor: editor.mode === 'yaml' ? '#007bff' : '#f0f0f0',
              color: editor.mode === 'yaml' ? '#fff' : '#000',
              border: '1px solid #ddd',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            YAML
          </button>
          <button
            onClick={() => handleModeChange('sdk')}
            style={{
              padding: '8px 16px',
              backgroundColor: editor.mode === 'sdk' ? '#007bff' : '#f0f0f0',
              color: editor.mode === 'sdk' ? '#fff' : '#000',
              border: '1px solid #ddd',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Python SDK
          </button>
        </div>
      </div>

      {/* Editor Content */}
      <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
        <div style={{ flex: 1 }}>
          {editor.mode === 'dag' && <DAGEditor pipeline={editor.pipeline} onSelect={editor.selectNode} />}
          {editor.mode === 'yaml' && (
            <YAMLEditor
              initialYAML={yamlContent}
              onChange={async (yaml) => {
                try {
                  await editor.importPipeline(yaml, 'yaml');
                } catch (err) {
                  console.error('Import error:', err);
                }
              }}
            />
          )}
          {editor.mode === 'sdk' && (
            <SDKEditor
              initialPython={pythonContent}
              onChange={async (py) => {
                try {
                  await editor.importPipeline(py, 'py');
                } catch (err) {
                  console.error('Import error:', err);
                }
              }}
            />
          )}
        </div>

        {/* Right Panel */}
        <div
          style={{
            width: '300px',
            backgroundColor: '#f9f9f9',
            borderLeft: '1px solid #ddd',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <div style={{ padding: '15px', borderBottom: '1px solid #ddd' }}>
            <h4 style={{ margin: '0 0 10px 0' }}>Actions</h4>
            <button
              onClick={() => editor.validate()}
              style={{
                display: 'block',
                width: '100%',
                padding: '8px',
                marginBottom: '8px',
                backgroundColor: '#28a745',
                color: '#fff',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
              }}
            >
              Validate
            </button>
            <button
              onClick={() => handleCompile('argo')}
              style={{
                display: 'block',
                width: '100%',
                padding: '8px',
                marginBottom: '8px',
                backgroundColor: '#17a2b8',
                color: '#fff',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
              }}
            >
              Compile to Argo
            </button>
            <button
              onClick={() => handleCompile('airflow')}
              style={{
                display: 'block',
                width: '100%',
                padding: '8px',
                backgroundColor: '#ffc107',
                color: '#000',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
              }}
            >
              Compile to Airflow
            </button>
          </div>

          {/* Validation Errors */}
          {editor.validationErrors.length > 0 && (
            <div style={{ padding: '15px', backgroundColor: '#f8d7da', borderBottom: '1px solid #ddd' }}>
              <h5 style={{ margin: '0 0 8px 0', color: '#721c24' }}>Validation Errors ({editor.validationErrors.length})</h5>
              <ul style={{ margin: 0, paddingLeft: '20px', fontSize: '12px' }}>
                {editor.validationErrors.map((err, i) => (
                  <li key={i} style={{ color: '#721c24' }}>
                    {err.message}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Compilation Result */}
          {editor.compilationResult && (
            <div style={{ padding: '15px', backgroundColor: '#d4edda', borderBottom: '1px solid #ddd' }}>
              <h5 style={{ margin: '0 0 8px 0', color: '#155724' }}>Compilation Successful</h5>
              <p style={{ margin: 0, fontSize: '12px', color: '#155724' }}>
                Ready to deploy to {editor.compilationResult.executor}
              </p>
            </div>
          )}

          {/* Status */}
          <div style={{ flex: 1, padding: '15px', fontSize: '12px', color: '#666' }}>
            {editor.isDirty && <p>✏️ Unsaved changes</p>}
            {editor.isLoading && <p>⏳ Loading...</p>}
            {editor.error && <p style={{ color: '#d32f2f' }}>❌ {editor.error}</p>}
          </div>
        </div>
      </div>
    </div>
  );
};
