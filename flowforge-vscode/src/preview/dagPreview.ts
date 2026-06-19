import * as vscode from 'vscode';
import { PipelineSpec } from '../parser/pythonParser';

function renderNodeList(spec: PipelineSpec): string {
  return spec.tasks
    .map(t => `<div style="padding:8px 10px;border:1px solid #444;border-radius:8px;margin-bottom:8px;background:#1f1f1f;color:#e6e6e6;">${t.id}</div>`)
    .join('');
}

function renderEdgeList(spec: PipelineSpec): string {
  if (spec.edges.length === 0) {
    return '<div style="color:#ccc;">No edges defined.</div>';
  }
  return spec.edges
    .map(e => `<div style="margin-bottom:6px;color:#9cdcfe;">${e.from} → ${e.to}</div>`)
    .join('');
}

export function showDagPreview(spec: PipelineSpec): void {
  const panel = vscode.window.createWebviewPanel(
    'flowforgeDagPreview',
    `FlowForge DAG Preview: ${spec.meta.name}`,
    vscode.ViewColumn.Beside,
    { enableScripts: false }
  );

  panel.webview.html = `<!DOCTYPE html>
<html>
<body style="font-family:Segoe UI, Arial, sans-serif;background:#111;padding:16px;">
  <h2 style="color:#fff;margin-top:0;">Pipeline: ${spec.meta.name}</h2>
  <div style="color:#ccc;margin-bottom:14px;">Version: ${spec.meta.version || 'n/a'} | Owner: ${spec.meta.owner || 'unknown'}</div>
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;">
    <div>
      <h3 style="color:#fff;">Tasks</h3>
      ${renderNodeList(spec)}
    </div>
    <div>
      <h3 style="color:#fff;">Dependencies</h3>
      ${renderEdgeList(spec)}
    </div>
  </div>
</body>
</html>`;
}
