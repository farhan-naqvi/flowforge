"use strict";
// Covenant workbench UI. All contract/verdict/log data is inserted via
// textContent (never innerHTML), so contract or sample values cannot inject markup.

const state = { slug: null, data: null };

function el(tag, cls, text) {
  const n = document.createElement(tag);
  if (cls) n.className = cls;
  if (text !== undefined) n.textContent = text;
  return n;
}
async function api(path, opts) {
  const r = await fetch(path, opts);
  if (!r.ok) { let m = r.statusText; try { m = (await r.json()).error || m; } catch (e) {} throw new Error(m); }
  return r.json();
}
function setConn(ok, text) {
  document.getElementById("dot").className = "dot " + (ok ? "ok" : "err");
  document.getElementById("conn-text").textContent = text;
}
function badge(t, k) { return el("span", "badge " + k, t); }

async function loadProducts() {
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
      if (p.slug === state.slug) li.classList.add("active");
      list.appendChild(li);
    });
  } catch (e) { setConn(false, "disconnected"); }
}

async function selectProduct(slug, li) {
  document.querySelectorAll(".product-list li").forEach((n) => n.classList.remove("active"));
  if (li) li.classList.add("active");
  state.slug = slug;
  document.getElementById("placeholder").classList.add("hidden");
  document.getElementById("product").classList.remove("hidden");
  document.getElementById("verdict").textContent = "";
  document.getElementById("artifact-card").classList.add("hidden");
  try { state.data = await api("/api/product?slug=" + encodeURIComponent(slug)); render(); }
  catch (e) { document.getElementById("verdict").appendChild(el("div", "verdict-line err", e.message)); }
}

function schemaTable(tableEl, schema, pk) {
  tableEl.textContent = "";
  schema.forEach((f) => {
    const tr = el("tr");
    tr.appendChild(el("td", null, f.name));
    tr.appendChild(el("td", "t", f.type));
    const flags = el("td");
    if ((pk || []).includes(f.name)) flags.appendChild(el("span", "pk", "PK "));
    if (!f.nullable) flags.appendChild(el("span", "req", "required"));
    tr.appendChild(flags);
    tableEl.appendChild(tr);
  });
}

function render() {
  const d = state.data;
  document.getElementById("p-slug").textContent = d.slug;
  document.getElementById("src-id").textContent = d.source.id;
  document.getElementById("tgt-id").textContent = d.target.id;
  schemaTable(document.getElementById("src-schema"), d.source.schema, d.source.primary_key);
  schemaTable(document.getElementById("tgt-schema"), d.target.schema, d.target.primary_key);

  const steps = document.getElementById("steps");
  steps.textContent = "";
  d.plan.steps.forEach((s) => {
    const li = el("li");
    li.appendChild(el("span", "prim", s.primitive));
    if (s.output_schema) li.appendChild(el("span", "arrow", "  →  " + s.output_schema.map((f) => f.name).join(", ")));
    steps.appendChild(li);
  });

  const conf = document.getElementById("static-conf");
  const ok = d.plan.conformance.ok;
  conf.className = "conf " + (ok ? "ok" : "err");
  conf.textContent = ok
    ? "✓ static: schema inference conforms to the target contract (before any data moves)"
    : "✗ static: " + d.plan.conformance.problems.join("; ");

  const badges = document.getElementById("p-badges");
  badges.textContent = "";
  badges.appendChild(badge(ok ? "conforms" : "non-conformant", ok ? "ok" : "err"));

  document.getElementById("verify-btn").disabled = !ok;
}

async function verify() {
  const box = document.getElementById("verdict");
  box.textContent = "";
  const btn = document.getElementById("verify-btn");
  btn.disabled = true; btn.textContent = "Verifying…";
  try {
    const res = await api("/api/verify", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ slug: state.slug }),
    });
    const s = res.static, dyn = res.dynamic;
    box.appendChild(line(s.ok, s.ok ? "static: schema inference conforms" : "static: " + s.problems.join("; ")));
    const dl = line(dyn.ok, `dynamic: local DuckDB run on synthetic + edge-case data — ${dyn.row_count} rows`);
    box.appendChild(dl);
    (dyn.schema_problems || []).concat(dyn.quality_problems || []).forEach((p) => box.appendChild(el("div", "problem", "• " + p)));
    if (dyn.sample && dyn.sample.length) {
      const pre = el("div", "sample", dyn.sample.slice(0, 3).map((r) => JSON.stringify(r)).join("\n"));
      box.appendChild(pre);
    }
    if (dyn.ok) box.appendChild(el("div", "verdict-line ok", "✓ trusted to run at full scale — the green check"));
  } catch (e) { box.appendChild(el("div", "verdict-line err", e.message)); }
  finally { btn.disabled = false; btn.textContent = "Verify against contract"; }
}

function line(ok, text) {
  const d = el("div", "verdict-line " + (ok ? "ok" : "err"));
  d.appendChild(el("span", null, ok ? "✓" : "✗"));
  d.appendChild(el("span", null, text));
  return d;
}

async function compile() {
  try {
    const res = await api("/api/compile?slug=" + encodeURIComponent(state.slug));
    const card = document.getElementById("artifact-card");
    card.classList.remove("hidden");
    document.getElementById("artifact-title").textContent = "Argo/Spark artifact";
    document.getElementById("artifact").textContent = res.ok ? res.artifact
      : "Cannot compile: plan does not conform to the target contract.";
  } catch (e) { alert(e.message); }
}

async function savePlan() {
  try {
    const res = await api("/api/plan/save", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ slug: state.slug }),
    });
    const box = document.getElementById("verdict");
    box.appendChild(el("div", "verdict-line " + (res.ok ? "ok" : "err"),
      res.ok ? `wrote ${res.path} — ${res.note}` : res.reason));
  } catch (e) { alert(e.message); }
}

document.getElementById("verify-btn").addEventListener("click", verify);
document.getElementById("compile-btn").addEventListener("click", compile);
document.getElementById("save-btn").addEventListener("click", savePlan);
document.getElementById("copy-btn").addEventListener("click", () => {
  const t = document.getElementById("artifact").textContent;
  if (navigator.clipboard) navigator.clipboard.writeText(t);
});
loadProducts();
setInterval(() => { if (document.getElementById("dot").className.indexOf("err") >= 0) loadProducts(); }, 5000);
