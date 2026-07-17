"use strict";
// Covenant workbench UI. All contract/verdict/sample data is inserted via
// textContent or input .value (never innerHTML), so authored or sample values
// cannot inject markup.

const state = {
  slug: null,
  data: null,
  meta: { dtypes: [], primitives: [] },
  src: [], tgt: [], steps: [],   // editable working copies
  dirty: false,
};

function el(tag, cls, text) {
  const n = document.createElement(tag);
  if (cls) n.className = cls;
  if (text !== undefined) n.textContent = text;
  return n;
}
async function api(path, opts) {
  const r = await fetch(path, opts);
  const body = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(body.error || r.statusText);
  return body;
}
function post(path, obj) {
  return api(path, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(obj) });
}
function setConn(ok, text) {
  document.getElementById("dot").className = "dot " + (ok ? "ok" : "err");
  document.getElementById("conn-text").textContent = text;
}
function badge(t, k) { return el("span", "badge " + k, t); }
function showErr(node, msg) { node.textContent = msg; node.classList.remove("hidden"); }
function clearErr(node) { node.textContent = ""; node.classList.add("hidden"); }

// -- boot -----------------------------------------------------------------
async function boot() {
  try { state.meta = await api("/api/meta"); } catch (e) {}
  await loadProducts();
}

async function loadProducts(selectSlug) {
  try {
    const { products } = await api("/api/products");
    setConn(true, "connected");
    const list = document.getElementById("product-list");
    list.textContent = "";
    document.getElementById("empty").classList.toggle("hidden", products.length > 0);
    products.forEach((p) => {
      const li = el("li", null, p.slug);
      li.setAttribute("role", "button"); li.setAttribute("tabindex", "0");
      const go = () => selectProduct(p.slug, li);
      li.addEventListener("click", go);
      li.addEventListener("keydown", (e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); go(); } });
      if (p.slug === (selectSlug || state.slug)) li.classList.add("active");
      list.appendChild(li);
    });
    if (selectSlug) selectProduct(selectSlug);
  } catch (e) { setConn(false, "disconnected"); }
}

async function selectProduct(slug, li) {
  document.querySelectorAll(".product-list li").forEach((n) => n.classList.remove("active"));
  if (li) li.classList.add("active"); else
    document.querySelectorAll(".product-list li").forEach((n) => { if (n.textContent === slug) n.classList.add("active"); });
  state.slug = slug;
  document.getElementById("placeholder").classList.add("hidden");
  document.getElementById("product").classList.remove("hidden");
  clearVerdict();
  try { setData(await api("/api/product?slug=" + encodeURIComponent(slug))); }
  catch (e) { showVerdictError(e.message); }
}

function setData(product) {
  state.data = product;
  // clear any stale per-section error/flash messages from a previous product
  document.querySelectorAll('[data-err="source"], [data-err="target"], #intent-error')
    .forEach((n) => { n.textContent = ""; n.className = "err-inline hidden"; });
  // hydrate editable working copies
  const pk = (kind) => new Set(product[kind].primary_key || []);
  const mkFields = (kind) => product[kind].schema.map((f) => ({
    name: f.name, type: f.type, nullable: f.nullable, pk: pk(kind).has(f.name),
  }));
  state.src = mkFields("source");
  state.tgt = mkFields("target");
  state.steps = (product.intent || []).map((s) => ({
    primitive: s.primitive, paramsText: JSON.stringify(s.params || {}, null, 2),
  }));
  state.dirty = false;
  render();
}

// -- render ---------------------------------------------------------------
function render() {
  const d = state.data;
  document.getElementById("p-slug").textContent = d.slug;
  document.getElementById("src-id").textContent = d.source.id;
  document.getElementById("tgt-id").textContent = d.target.id;
  renderContractEditor("source", "src-editor", state.src);
  renderContractEditor("target", "tgt-editor", state.tgt);
  renderIntentEditor();
  renderPlan();
  renderBadges();
}

function typeSelect(value, onChange) {
  const sel = el("select");
  state.meta.dtypes.forEach((t) => {
    const o = el("option", null, t); o.value = t; if (t === value) o.selected = true; sel.appendChild(o);
  });
  sel.addEventListener("change", () => { onChange(sel.value); markDirty(); });
  return sel;
}
function checkbox(checked, onChange) {
  const c = el("input"); c.type = "checkbox"; c.checked = !!checked;
  c.addEventListener("change", () => { onChange(c.checked); markDirty(); });
  return c;
}

function renderContractEditor(kind, tableId, fields) {
  const table = document.getElementById(tableId);
  table.textContent = "";
  fields.forEach((f, i) => {
    const tr = el("tr");
    // name
    const tdName = el("td");
    const name = el("input"); name.type = "text"; name.value = f.name; name.setAttribute("aria-label", "field name");
    name.addEventListener("input", () => { f.name = name.value; markDirty(); });
    tdName.appendChild(name); tr.appendChild(tdName);
    // type
    const tdType = el("td");
    tdType.appendChild(typeSelect(f.type, (v) => (f.type = v))); tr.appendChild(tdType);
    // nullable
    const tdNull = el("td", "flag");
    const nl = el("label"); nl.appendChild(checkbox(f.nullable, (v) => (f.nullable = v))); nl.appendChild(el("span", null, "nullable"));
    tdNull.appendChild(nl); tr.appendChild(tdNull);
    // PK
    const tdPk = el("td", "flag");
    const pl = el("label"); pl.appendChild(checkbox(f.pk, (v) => (f.pk = v))); pl.appendChild(el("span", null, "PK"));
    tdPk.appendChild(pl); tr.appendChild(tdPk);
    // ordering + remove
    const tdOrd = el("td", "ord");
    tdOrd.appendChild(moveBtn("↑", () => moveItem(fields, i, -1, () => renderContractEditor(kind, tableId, fields))));
    tdOrd.appendChild(moveBtn("↓", () => moveItem(fields, i, +1, () => renderContractEditor(kind, tableId, fields))));
    tdOrd.appendChild(delBtn(() => { fields.splice(i, 1); markDirty(); renderContractEditor(kind, tableId, fields); }));
    tr.appendChild(tdOrd);
    table.appendChild(tr);
  });
}

function renderIntentEditor() {
  const box = document.getElementById("intent-editor");
  box.textContent = "";
  if (state.steps.length === 0) box.appendChild(el("p", "muted", "No steps yet. Add transform steps to build the pipeline."));
  state.steps.forEach((s, i) => {
    const step = el("div", "step");
    const head = el("div", "step-head");
    head.appendChild(el("span", "num", "#" + (i + 1)));
    const sel = el("select");
    state.meta.primitives.filter((p) => p !== "source" && p !== "sink").forEach((p) => {
      const o = el("option", null, p); o.value = p; if (p === s.primitive) o.selected = true; sel.appendChild(o);
    });
    sel.addEventListener("change", () => { s.primitive = sel.value; markDirty(); });
    head.appendChild(sel);
    head.appendChild(el("span", "spacer"));
    head.appendChild(moveBtn("↑", () => moveItem(state.steps, i, -1, renderIntentEditor)));
    head.appendChild(moveBtn("↓", () => moveItem(state.steps, i, +1, renderIntentEditor)));
    head.appendChild(delBtn(() => { state.steps.splice(i, 1); markDirty(); renderIntentEditor(); }));
    step.appendChild(head);
    step.appendChild(el("label", "plabel", "params (JSON)"));
    const ta = el("textarea"); ta.value = s.paramsText; ta.spellcheck = false;
    ta.setAttribute("aria-label", "step params JSON");
    ta.addEventListener("input", () => { s.paramsText = ta.value; markDirty(); });
    step.appendChild(ta);
    box.appendChild(step);
  });
}

function renderPlan() {
  const steps = document.getElementById("steps");
  const conf = document.getElementById("static-conf");
  steps.textContent = "";
  if (!state.data.plan) {
    conf.className = "conf err";
    conf.textContent = "✗ " + (state.data.plan_error || "plan could not be built");
    document.getElementById("verify-btn").disabled = true;
    return;
  }
  state.data.plan.steps.forEach((s) => {
    const li = el("li");
    li.appendChild(el("span", "prim", s.primitive));
    if (s.output_schema) li.appendChild(el("span", "arrow", "  →  " + s.output_schema.map((f) => f.name).join(", ")));
    steps.appendChild(li);
  });
  const ok = state.data.plan.conformance.ok;
  conf.className = "conf " + (ok ? "ok" : "err");
  conf.textContent = ok
    ? "✓ static: schema inference conforms to the target contract (before any data moves)"
    : "✗ static: " + state.data.plan.conformance.problems.join("; ");
  document.getElementById("verify-btn").disabled = !ok;
}

function renderBadges() {
  const box = document.getElementById("p-badges");
  box.textContent = "";
  if (state.dirty) box.appendChild(badge("unsaved changes", "warn"));
  if (!state.data.plan) { box.appendChild(badge("plan error", "err")); return; }
  box.appendChild(badge(state.data.plan.conformance.ok ? "conforms" : "non-conformant", state.data.plan.conformance.ok ? "ok" : "err"));
}

// -- small controls -------------------------------------------------------
function moveBtn(sym, fn) { const b = el("button", "iconbtn", sym); b.type = "button"; b.addEventListener("click", fn); return b; }
function delBtn(fn) { const b = el("button", "iconbtn danger", "✕"); b.type = "button"; b.title = "remove"; b.addEventListener("click", fn); return b; }
function moveItem(arr, i, delta, rerender) {
  const j = i + delta;
  if (j < 0 || j >= arr.length) return;
  [arr[i], arr[j]] = [arr[j], arr[i]];
  markDirty(); rerender();
}
function markDirty() { state.dirty = true; renderBadges(); }

// -- saves ----------------------------------------------------------------
async function saveContract(kind) {
  const errNode = document.querySelector(`[data-err="${kind}"]`);
  clearErr(errNode);
  const fields = (kind === "source" ? state.src : state.tgt).map((f) => ({
    name: f.name, type: f.type, nullable: f.nullable, primary_key: f.pk,
  }));
  try {
    const res = await post("/api/contract/save", { slug: state.slug, kind, fields });
    setData(res.product);              // re-planned; conformance refreshes
    clearVerdict();
    flash(errNode, "saved ✓ " + res.path);
  } catch (e) { showErr(errNode, e.message); }
}

async function saveIntent() {
  const errNode = document.getElementById("intent-error");
  clearErr(errNode);
  const steps = [];
  for (let i = 0; i < state.steps.length; i++) {
    const s = state.steps[i];
    let params;
    try { params = s.paramsText.trim() ? JSON.parse(s.paramsText) : {}; }
    catch (e) { return showErr(errNode, `step #${i + 1} (${s.primitive}): params is not valid JSON`); }
    steps.push({ primitive: s.primitive, params });
  }
  try {
    const res = await post("/api/intent/save", { slug: state.slug, steps });
    setData(res.product);
    clearVerdict();
    flash(errNode, "saved ✓ " + res.path);
  } catch (e) { showErr(errNode, e.message); }
}

function flash(node, msg) {
  node.textContent = msg; node.className = "save-ok"; node.classList.remove("hidden");
  setTimeout(() => { node.classList.add("hidden"); node.className = "err-inline hidden"; }, 2500);
}

// -- new product ----------------------------------------------------------
function toggleNewForm(show) {
  const f = document.getElementById("new-form");
  f.classList.toggle("hidden", !show);
  if (show) { document.getElementById("new-slug").value = ""; clearErr(document.getElementById("new-error")); document.getElementById("new-slug").focus(); }
}
async function createProduct(e) {
  e.preventDefault();
  const slug = document.getElementById("new-slug").value.trim();
  const errNode = document.getElementById("new-error");
  clearErr(errNode);
  try {
    await post("/api/product/create", { slug });
    toggleNewForm(false);
    await loadProducts(slug);          // reload list and open the new product
  } catch (e2) { showErr(errNode, e2.message); }
}

// -- verify / compile / plan-save (unchanged pipeline) --------------------
function clearVerdict() {
  document.getElementById("verdict").textContent = "";
  document.getElementById("artifact-card").classList.add("hidden");
}
function showVerdictError(msg) { document.getElementById("verdict").appendChild(el("div", "verdict-line err", msg)); }
function line(ok, text) {
  const d = el("div", "verdict-line " + (ok ? "ok" : "err"));
  d.appendChild(el("span", null, ok ? "✓" : "✗")); d.appendChild(el("span", null, text)); return d;
}
async function verify() {
  const box = document.getElementById("verdict"); box.textContent = "";
  if (state.dirty) box.appendChild(el("div", "verdict-line err", "You have unsaved changes — save the contract/intent first."));
  const btn = document.getElementById("verify-btn"); btn.disabled = true; btn.textContent = "Verifying…";
  try {
    const res = await post("/api/verify", { slug: state.slug });
    const s = res.static, dyn = res.dynamic;
    box.appendChild(line(s.ok, s.ok ? "static: schema inference conforms" : "static: " + s.problems.join("; ")));
    box.appendChild(line(dyn.ok, `dynamic: local DuckDB run on synthetic + edge-case data — ${dyn.row_count} rows`));
    (dyn.schema_problems || []).concat(dyn.quality_problems || []).forEach((p) => box.appendChild(el("div", "problem", "• " + p)));
    if (dyn.sample && dyn.sample.length) box.appendChild(el("div", "sample", dyn.sample.slice(0, 3).map((r) => JSON.stringify(r)).join("\n")));
    if (dyn.ok) box.appendChild(el("div", "verdict-line ok", "✓ trusted to run at full scale — the green check"));
  } catch (e) { showVerdictError(e.message); }
  finally { btn.disabled = false; btn.textContent = "Verify against contract"; }
}
async function compile() {
  try {
    const res = await api("/api/compile?slug=" + encodeURIComponent(state.slug));
    const card = document.getElementById("artifact-card"); card.classList.remove("hidden");
    document.getElementById("artifact-title").textContent = "Argo/Spark artifact";
    document.getElementById("artifact").textContent = res.ok ? res.artifact
      : "Cannot compile: plan does not conform to the target contract.";
  } catch (e) { showVerdictError(e.message); }
}
async function savePlan() {
  try {
    const res = await post("/api/plan/save", { slug: state.slug });
    document.getElementById("verdict").appendChild(el("div", "verdict-line " + (res.ok ? "ok" : "err"),
      res.ok ? `wrote ${res.path} — ${res.note}` : res.reason));
  } catch (e) { showVerdictError(e.message); }
}

// -- wiring ---------------------------------------------------------------
document.getElementById("new-btn").addEventListener("click", () => toggleNewForm(true));
document.getElementById("new-cancel").addEventListener("click", () => toggleNewForm(false));
document.getElementById("new-form").addEventListener("submit", createProduct);
document.querySelectorAll("[data-add]").forEach((b) => b.addEventListener("click", () => {
  const kind = b.dataset.add;
  (kind === "source" ? state.src : state.tgt).push({ name: "new_field", type: "string", nullable: true, pk: false });
  markDirty();
  renderContractEditor(kind, kind === "source" ? "src-editor" : "tgt-editor", kind === "source" ? state.src : state.tgt);
}));
document.querySelectorAll("[data-save-contract]").forEach((b) => b.addEventListener("click", () => saveContract(b.dataset.saveContract)));
document.getElementById("add-step").addEventListener("click", () => {
  state.steps.push({ primitive: (state.meta.primitives.find((p) => p !== "source" && p !== "sink") || "filter"), paramsText: "{}" });
  markDirty(); renderIntentEditor();
});
document.getElementById("save-intent").addEventListener("click", saveIntent);
document.getElementById("verify-btn").addEventListener("click", verify);
document.getElementById("compile-btn").addEventListener("click", compile);
document.getElementById("save-btn").addEventListener("click", savePlan);
document.getElementById("copy-btn").addEventListener("click", () => {
  const t = document.getElementById("artifact").textContent;
  if (navigator.clipboard) navigator.clipboard.writeText(t);
});
boot();
setInterval(() => { if (document.getElementById("dot").className.indexOf("err") >= 0) loadProducts(); }, 5000);
