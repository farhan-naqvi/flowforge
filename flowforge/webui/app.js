"use strict";
// FlowForge workbench UI. All user/pipeline data is inserted via textContent
// (never innerHTML) so poisoned logs or names cannot inject markup.

const state = { file: null, data: null, target: "argo", artifact: null };

function el(tag, cls, text) {
  const n = document.createElement(tag);
  if (cls) n.className = cls;
  if (text !== undefined) n.textContent = text;
  return n;
}

async function api(path, opts) {
  const res = await fetch(path, opts);
  if (!res.ok) {
    let msg = res.statusText;
    try { msg = (await res.json()).error || msg; } catch (e) {}
    throw new Error(msg);
  }
  return res.json();
}

function setConn(ok, text) {
  document.getElementById("conn-dot").className = "dot " + (ok ? "ok" : "err");
  document.getElementById("conn-text").textContent = text;
}

async function loadPipelines() {
  try {
    const { pipelines } = await api("/api/pipelines");
    setConn(true, "connected");
    const list = document.getElementById("pipeline-list");
    list.textContent = "";
    document.getElementById("pipeline-empty").classList.toggle("hidden", pipelines.length > 0);
    pipelines.forEach((p) => {
      const li = el("li");
      li.setAttribute("role", "button");
      li.setAttribute("tabindex", "0");
      li.appendChild(el("div", "p-name", p.name));
      li.appendChild(el("div", "p-sub", `${p.file} · ${p.tasks} tasks, ${p.edges} edges`));
      const activate = () => selectPipeline(p.file, li);
      li.addEventListener("click", activate);
      li.addEventListener("keydown", (e) => {
        if (e.key === "Enter" || e.key === " ") { e.preventDefault(); activate(); }
      });
      if (p.file === state.file) li.classList.add("active");
      list.appendChild(li);
    });
  } catch (e) {
    setConn(false, "disconnected");
  }
}

async function selectPipeline(file, li) {
  document.querySelectorAll(".pipeline-list li").forEach((n) => n.classList.remove("active"));
  if (li) li.classList.add("active");
  state.file = file;
  state.artifact = null;
  document.getElementById("placeholder").classList.add("hidden");
  document.getElementById("pipeline-detail").classList.remove("hidden");
  try {
    state.data = await api("/api/pipeline?file=" + encodeURIComponent(file));
    renderDetail();
  } catch (e) {
    showBanner("Failed to load pipeline: " + e.message);
  }
}

function renderDetail() {
  const d = state.data;
  const meta = d.ir.metadata;
  document.getElementById("pd-name").textContent = meta.name;
  const bits = [];
  if (meta.version) bits.push("v" + meta.version);
  if (meta.owner) bits.push(meta.owner);
  bits.push(Object.keys(d.ir.tasks).length + " tasks");
  document.getElementById("pd-meta").textContent = bits.join(" · ");

  renderStateBadges();
  renderDiagnostics();
  document.getElementById("ir-body").textContent = JSON.stringify(d.ir, null, 2);
  renderTargets();
  loadRuns();
}

function renderStateBadges() {
  const box = document.getElementById("pd-state");
  box.textContent = "";
  const v = state.data.validation;
  if (v.ok) {
    box.appendChild(badge("validated", "ok"));
  } else {
    box.appendChild(badge(plural(v.errors.length, "error"), "err"));
  }
  if (v.warnings && v.warnings.length) box.appendChild(badge(plural(v.warnings.length, "warning"), "warn"));
}

function badge(text, kind) { return el("span", "badge " + kind, text); }

function plural(n, word) { return n + " " + word + (n === 1 ? "" : "s"); }

function renderDiagnostics() {
  const body = document.getElementById("diag-body");
  body.textContent = "";
  const v = state.data.validation;
  const all = v.errors.concat(v.warnings);
  if (all.length === 0) {
    body.appendChild(el("div", "diag-ok", "✓ No problems found. Graph is acyclic, identifiers are injection-safe, schema conforms."));
    return;
  }
  all.forEach((diag) => {
    const item = el("div", "diag-item " + diag.severity);
    item.appendChild(el("span", "mark", diag.severity === "error" ? "✗" : "!"));
    const main = el("div");
    main.appendChild(el("div", null, diag.message));
    const sub = el("div", "diag-code", diag.code + (diag.location ? "  ·  " + diag.location : ""));
    main.appendChild(sub);
    item.appendChild(main);
    body.appendChild(item);
  });
}

function renderTargets() {
  const sw = document.getElementById("target-switch");
  sw.textContent = "";
  Object.keys(state.data.capability).forEach((t) => {
    const b = el("button", "btn" + (t === state.target ? " active" : ""), t);
    b.addEventListener("click", () => { state.target = t; renderTargets(); compileTarget(); });
    sw.appendChild(b);
  });
  renderCapability();
  compileTarget();
}

function renderCapability() {
  const rep = state.data.capability[state.target];
  const body = document.getElementById("cap-body");
  body.textContent = "";
  rep.items.forEach((it) => {
    const tr = el("tr", it.used ? "" : "unused");
    tr.appendChild(el("td", "feat", it.feature));
    const lvl = el("td");
    lvl.appendChild(el("span", "lvl " + it.level, it.level));
    tr.appendChild(lvl);
    tr.appendChild(el("td", null, it.note + (it.used ? "" : " (not used)")));
    body.appendChild(tr);
  });
}

async function compileTarget() {
  const label = document.getElementById("artifact-label");
  const out = document.getElementById("artifact-body");
  label.textContent = state.target + " artifact";
  out.textContent = "compiling…";
  try {
    const res = await api(`/api/compile?file=${encodeURIComponent(state.file)}&target=${state.target}`);
    if (!res.ok) {
      out.textContent = "Cannot compile: pipeline has validation errors (see Diagnostics tab).";
      state.artifact = null;
      return;
    }
    state.artifact = res.artifact;
    out.textContent = res.artifact;
  } catch (e) {
    out.textContent = "Error: " + e.message;
  }
}

async function loadRuns() {
  const list = document.getElementById("run-list");
  list.textContent = "";
  try {
    const { runs } = await api("/api/runs");
    const mine = runs.filter((r) => r.pipeline === state.data.ir.metadata.name);
    if (mine.length === 0) { list.appendChild(el("li", "muted", "No local runs yet.")); return; }
    mine.forEach((r) => {
      const li = el("li");
      li.appendChild(el("span", "task-name", r.id));
      const right = el("span");
      right.appendChild(badge(r.status, r.status === "succeeded" ? "ok" : r.status === "failed" ? "err" : "neutral"));
      right.appendChild(el("span", "badge neutral", r.mode === "local" ? "LOCAL RUN" : r.mode));
      li.appendChild(right);
      li.addEventListener("click", () => showRun(r.id));
      list.appendChild(li);
    });
  } catch (e) { list.appendChild(el("li", "muted", "Could not load runs.")); }
}

async function runLocal() {
  const btn = document.getElementById("run-btn");
  btn.disabled = true; btn.textContent = "Running…";
  const cur = document.getElementById("run-current");
  cur.textContent = "";
  try {
    const res = await api("/api/run", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ file: state.file }),
    });
    renderRunCard(res);
    loadRuns();
  } catch (e) {
    showBanner("Run failed: " + e.message);
  } finally {
    btn.disabled = false; btn.textContent = "Run locally";
  }
}

async function showRun(id) {
  try {
    const run = await api("/api/run?id=" + encodeURIComponent(id));
    // normalise store shape to run-card shape
    renderRunCard({
      run_id: run.id, status: run.status,
      tasks: run.tasks.map((t) => ({ task: t.task, status: t.status, duration_ms: t.duration_ms, logs: t.logs, error: t.error })),
      validation_errors: [],
    });
  } catch (e) { showBanner(e.message); }
}

function renderRunCard(res) {
  const cur = document.getElementById("run-current");
  cur.textContent = "";
  const card = el("div", "run-card");
  const head = el("div", "task-row");
  head.appendChild(el("span", "task-name", "run " + res.run_id));
  head.appendChild(badge(res.status, res.status === "succeeded" ? "ok" : res.status === "failed" ? "err" : "warn"));
  head.appendChild(el("span", "badge neutral", "LOCAL RUN"));
  card.appendChild(head);
  if (res.status === "invalid") {
    (res.validation_errors || []).forEach((m) => card.appendChild(el("div", "task-err", m)));
  }
  (res.tasks || []).forEach((t) => {
    const row = el("div", "task-row");
    row.appendChild(el("span", "task-status " + t.status, t.status === "succeeded" ? "✓" : t.status === "failed" ? "✗" : "○"));
    row.appendChild(el("span", "task-name", t.task));
    row.appendChild(el("span", "task-dur", t.duration_ms + " ms"));
    card.appendChild(row);
    if (t.logs && t.logs.trim()) card.appendChild(el("div", "task-logs", t.logs.trim()));
    if (t.error) card.appendChild(el("div", "task-err", t.error.split("\n")[0]));
  });
  cur.appendChild(card);
}

function showBanner(msg) {
  const cur = document.getElementById("run-current");
  const b = el("div", "banner err", msg);
  cur.prepend(b);
}

function download(name, text) {
  const blob = new Blob([text], { type: "text/plain" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url; a.download = name; a.click();
  URL.revokeObjectURL(url);
}

// -- wiring ---------------------------------------------------------------
document.querySelectorAll(".tab").forEach((tab) => {
  tab.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach((t) => t.classList.remove("active"));
    document.querySelectorAll(".panel").forEach((p) => p.classList.remove("active"));
    tab.classList.add("active");
    document.querySelector(`.panel[data-panel="${tab.dataset.tab}"]`).classList.add("active");
  });
});
document.querySelectorAll("[data-copy]").forEach((btn) => {
  btn.addEventListener("click", () => {
    const what = btn.dataset.copy;
    const text = what === "ir" ? document.getElementById("ir-body").textContent : (state.artifact || "");
    if (navigator.clipboard) navigator.clipboard.writeText(text);
    btn.textContent = "Copied";
    setTimeout(() => (btn.textContent = "Copy"), 1200);
  });
});
document.getElementById("download-artifact").addEventListener("click", () => {
  if (!state.artifact) return;
  const ext = state.target === "argo" ? ".yaml" : ".py";
  download(state.data.ir.metadata.name + "." + state.target + ext, state.artifact);
});
document.getElementById("run-btn").addEventListener("click", runLocal);
document.getElementById("refresh").addEventListener("click", loadPipelines);

loadPipelines();
setInterval(() => { if (document.getElementById("conn-dot").className.indexOf("err") >= 0) loadPipelines(); }, 5000);
