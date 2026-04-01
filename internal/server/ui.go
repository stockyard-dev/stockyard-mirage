package server

var dashboardHTML = []byte(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Stockyard Mirage</title>
<style>
:root{--bg:#1a1410;--surface:#241c15;--border:#3d2e1e;--rust:#c4622d;--leather:#8b5e3c;--cream:#f5e6c8;--muted:#7a6550;--text:#e8d5b0}
*{box-sizing:border-box;margin:0;padding:0}
body{background:var(--bg);color:var(--text);font-family:'JetBrains Mono',monospace,sans-serif;min-height:100vh}
header{background:var(--surface);border-bottom:1px solid var(--border);padding:1rem 2rem;display:flex;align-items:center;gap:1rem}
.logo{color:var(--rust);font-size:1.25rem;font-weight:700}
.badge{background:var(--rust);color:var(--cream);font-size:0.65rem;padding:0.2rem 0.5rem;border-radius:3px;font-weight:600;text-transform:uppercase}
main{max-width:1100px;margin:0 auto;padding:2rem}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:1rem;margin-bottom:2rem}
.stat{background:var(--surface);border:1px solid var(--border);border-radius:6px;padding:1.25rem;text-align:center}
.stat-value{font-size:1.75rem;font-weight:700;color:var(--rust)}
.stat-label{font-size:0.75rem;color:var(--muted);margin-top:0.25rem;text-transform:uppercase;letter-spacing:0.05em}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:1rem}
.card{background:var(--surface);border:1px solid var(--border);border-radius:6px;padding:1.5rem}
.card h2{font-size:0.85rem;color:var(--muted);text-transform:uppercase;letter-spacing:0.08em;margin-bottom:1rem}
.form-row{display:flex;gap:0.5rem;margin-bottom:0.75rem;flex-wrap:wrap}
select,input,textarea{background:var(--bg);border:1px solid var(--border);color:var(--text);padding:0.5rem 0.75rem;border-radius:4px;font-family:inherit;font-size:0.85rem}
textarea{width:100%;min-height:80px;resize:vertical}
input[type=number]{width:90px}
.btn{background:var(--rust);color:var(--cream);border:none;padding:0.5rem 1rem;border-radius:4px;cursor:pointer;font-family:inherit;font-size:0.85rem;font-weight:600}
.btn:hover{opacity:0.85}
.btn-sm{padding:0.25rem 0.6rem;font-size:0.75rem}
.btn-danger{background:#7a2020}
table{width:100%;border-collapse:collapse;font-size:0.82rem}
th{text-align:left;color:var(--muted);padding:0.5rem;border-bottom:1px solid var(--border);font-size:0.75rem;text-transform:uppercase}
td{padding:0.5rem;border-bottom:1px solid var(--border)}
.method{font-weight:700;color:var(--rust)}
.path{color:var(--cream);font-family:monospace}
.status{font-size:0.75rem;padding:0.15rem 0.4rem;border-radius:3px;background:var(--bg);border:1px solid var(--border)}
.empty{color:var(--muted);font-size:0.85rem;padding:1rem 0;text-align:center}
.full{grid-column:1/-1}
.mock-url{background:var(--bg);border:1px solid var(--border);border-radius:4px;padding:0.75rem;font-family:monospace;font-size:0.8rem;color:var(--leather);word-break:break-all}
</style>
</head>
<body>
<header>
  <span class="logo">&#x2B21; Stockyard</span>
  <span style="color:var(--muted)">/</span>
  <span style="color:var(--cream);font-weight:600">Mirage</span>
  <span class="badge">API Mock</span>
</header>
<main>
  <div class="stats">
    <div class="stat"><div class="stat-value" id="s-endpoints">0</div><div class="stat-label">Endpoints</div></div>
    <div class="stat"><div class="stat-value" id="s-logs">0</div><div class="stat-label">Requests Logged</div></div>
    <div class="stat"><div class="stat-value" id="s-tier">FREE</div><div class="stat-label">Tier</div></div>
  </div>
  <div class="grid">
    <div class="card">
      <h2>Add Endpoint</h2>
      <div class="form-row">
        <select id="f-method"><option>GET</option><option>POST</option><option>PUT</option><option>PATCH</option><option>DELETE</option><option>*</option></select>
        <input id="f-path" placeholder="/api/users" style="flex:1">
      </div>
      <div class="form-row">
        <input id="f-status" type="number" value="200">
        <input id="f-ct" placeholder="application/json" style="flex:1">
        <input id="f-delay" type="number" value="0" placeholder="delay ms">
      </div>
      <textarea id="f-body">{}</textarea>
      <div style="margin-top:0.75rem"><button class="btn" onclick="createEndpoint()">Add Endpoint</button></div>
    </div>
    <div class="card">
      <h2>Mock Base URL</h2>
      <p style="font-size:0.82rem;color:var(--muted);margin-bottom:0.75rem">Point your app at this prefix. Requests are matched to your configured endpoints.</p>
      <div class="mock-url" id="mock-url">loading...</div>
      <p style="font-size:0.75rem;color:var(--muted);margin-top:0.75rem">Example: GET /mock/api/users matches endpoint GET /api/users</p>
    </div>
    <div class="card full">
      <h2>Endpoints</h2>
      <table><thead><tr><th>Method</th><th>Path</th><th>Status</th><th>Content-Type</th><th>Delay</th><th></th></tr></thead>
      <tbody id="endpoints-body"><tr><td colspan="6" class="empty">No endpoints yet</td></tr></tbody></table>
    </div>
    <div class="card full">
      <h2>Request Log</h2>
      <table><thead><tr><th>Method</th><th>Path</th><th>Body</th><th>Time</th></tr></thead>
      <tbody id="logs-body"><tr><td colspan="4" class="empty">No requests yet</td></tr></tbody></table>
    </div>
  </div>
</main>
<script>
document.getElementById('mock-url').textContent = window.location.origin + '/mock';

async function api(method, path, body) {
  const opts = {method, headers:{'Content-Type':'application/json'}};
  if (body) opts.body = JSON.stringify(body);
  const r = await fetch(path, opts);
  return r.json();
}

async function load() {
  const [stats, limits] = await Promise.all([api('GET','/api/stats'), api('GET','/api/limits')]);
  document.getElementById('s-endpoints').textContent = stats.endpoints || 0;
  document.getElementById('s-logs').textContent = stats.logs || 0;
  document.getElementById('s-tier').textContent = (limits.tier||'free').toUpperCase();
  await loadEndpoints();
  await loadLogs();
}

async function loadEndpoints() {
  const data = await api('GET', '/api/endpoints');
  const tbody = document.getElementById('endpoints-body');
  if (!data || !data.length) { tbody.innerHTML = '<tr><td colspan="6" class="empty">No endpoints yet</td></tr>'; return; }
  tbody.innerHTML = data.map(function(e) { return '<tr><td><span class="method">'+e.method+'</span></td><td class="path">'+e.path+'</td><td><span class="status">'+e.status_code+'</span></td><td style="font-size:0.75rem;color:var(--muted)">'+e.content_type+'</td><td style="font-size:0.75rem;color:var(--muted)">'+e.delay_ms+'ms</td><td><button class="btn btn-sm btn-danger" onclick="deleteEndpoint('+e.id+')">Del</button></td></tr>'; }).join('');
}

async function loadLogs() {
  const data = await api('GET', '/api/logs?limit=50');
  const tbody = document.getElementById('logs-body');
  if (!data || !data.length) { tbody.innerHTML = '<tr><td colspan="4" class="empty">No requests yet</td></tr>'; return; }
  tbody.innerHTML = data.map(function(l) { return '<tr><td><span class="method">'+l.method+'</span></td><td class="path">'+l.path+'</td><td style="font-size:0.75rem;color:var(--muted)">'+String(l.request_body||'').substring(0,60)+'</td><td style="font-size:0.75rem;color:var(--muted)">'+new Date(l.responded_at).toLocaleTimeString()+'</td></tr>'; }).join('');
}

async function createEndpoint() {
  var e = {
    method: document.getElementById('f-method').value,
    path: document.getElementById('f-path').value.trim(),
    status_code: parseInt(document.getElementById('f-status').value)||200,
    content_type: document.getElementById('f-ct').value.trim()||'application/json',
    delay_ms: parseInt(document.getElementById('f-delay').value)||0,
    response_body: document.getElementById('f-body').value.trim(),
  };
  if (!e.path) { alert('Path required'); return; }
  var res = await api('POST', '/api/endpoints', e);
  if (res.error) { alert(res.error); return; }
  document.getElementById('f-path').value = '';
  document.getElementById('f-body').value = '{}';
  load();
}

async function deleteEndpoint(id) {
  if (!confirm('Delete endpoint?')) return;
  await api('DELETE', '/api/endpoints/'+id);
  load();
}

load();
setInterval(loadLogs, 5000);
</script>
</body>
</html>`)
