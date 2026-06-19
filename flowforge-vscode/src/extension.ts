import * as path from 'path';
import * as vscode from 'vscode';
import { compileToAirflow } from './compiler/airflowCompiler';
import { compileToArgo } from './compiler/argoCompiler';
import { parsePythonPipeline } from './parser/pythonParser';
import { showDagPreview } from './preview/dagPreview';
import { validatePipeline } from './validator/dagValidator';

const diagnostics = vscode.languages.createDiagnosticCollection('flowforge');

async function resolveTargetUri(uri?: vscode.Uri): Promise<vscode.Uri | undefined> {
  if (uri) {
    return uri;
  }
  const active = vscode.window.activeTextEditor?.document.uri;
  if (active?.fsPath.endsWith('.py')) {
    return active;
  }
  const picked = await vscode.window.showOpenDialog({
    canSelectMany: false,
    filters: { Python: ['py'] },
    openLabel: 'Select pipeline file'
  });
  return picked?.[0];
}

function pushProblems(uri: vscode.Uri, errors: string[]): void {
  if (errors.length === 0) {
    diagnostics.delete(uri);
    return;
  }
  const range = new vscode.Range(0, 0, 0, 1);
  diagnostics.set(
    uri,
    errors.map(msg => new vscode.Diagnostic(range, msg, vscode.DiagnosticSeverity.Error))
  );
}

async function generate(uri: vscode.Uri | undefined, kind: 'argo' | 'airflow'): Promise<void> {
  const target = await resolveTargetUri(uri);
  if (!target) {
    return;
  }

  const document = await vscode.workspace.openTextDocument(target);
  const source = document.getText();

  let spec;
  try {
    spec = parsePythonPipeline(source);
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    pushProblems(target, [message]);
    vscode.window.showErrorMessage(`FlowForge parse failed: ${message}`);
    return;
  }

  const issues = validatePipeline(spec);
  pushProblems(target, issues.map(i => i.message));
  if (issues.length > 0) {
    vscode.window.showErrorMessage('FlowForge validation failed. See Problems panel.');
    return;
  }

  const output = kind === 'argo' ? compileToArgo(spec) : compileToAirflow(spec);
  const extension = kind === 'argo' ? '.argo.yaml' : '.airflow.py';
  const outPath = path.join(path.dirname(target.fsPath), `${path.basename(target.fsPath, '.py')}${extension}`);
  const outUri = vscode.Uri.file(outPath);

  await vscode.workspace.fs.writeFile(outUri, Buffer.from(output, 'utf8'));
  const outDoc = await vscode.workspace.openTextDocument(outUri);
  await vscode.window.showTextDocument(outDoc, { preview: false });
  vscode.window.showInformationMessage(`Generated ${path.basename(outPath)}`);
}

export function activate(context: vscode.ExtensionContext): void {
  context.subscriptions.push(diagnostics);

  context.subscriptions.push(
    vscode.commands.registerCommand('flowforge.generateArgoWorkflow', (uri?: vscode.Uri) => generate(uri, 'argo'))
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('flowforge.generateAirflowDag', (uri?: vscode.Uri) => generate(uri, 'airflow'))
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('flowforge.previewDag', async (uri?: vscode.Uri) => {
      const target = await resolveTargetUri(uri);
      if (!target) {
        return;
      }
      const source = (await vscode.workspace.openTextDocument(target)).getText();
      try {
        const spec = parsePythonPipeline(source);
        const issues = validatePipeline(spec);
        pushProblems(target, issues.map(i => i.message));
        if (issues.length > 0) {
          vscode.window.showErrorMessage('FlowForge validation failed. See Problems panel.');
          return;
        }
        showDagPreview(spec);
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        pushProblems(target, [message]);
        vscode.window.showErrorMessage(`FlowForge parse failed: ${message}`);
      }
    })
  );
}

export function deactivate(): void {
  diagnostics.dispose();
}
