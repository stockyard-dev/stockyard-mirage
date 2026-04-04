package server
import "net/http"
func(s *Server)dashboard(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html");w.Write([]byte(dashHTML))}
const dashHTML=`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Mirage</title>
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}
.main{padding:1.5rem;max-width:900px;margin:0 auto}
.svc{background:var(--bg2);border:1px solid var(--bg3);margin-bottom:1rem}
.svc-hdr{padding:.8rem 1rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.svc-name{font-size:.85rem;color:var(--cream)}.svc-prefix{font-size:.65rem;color:var(--gold)}
.ep{display:flex;align-items:center;padding:.4rem 1rem;border-bottom:1px solid var(--bg3);font-size:.72rem}
.ep:last-child{border:none}
.ep-method{width:50px;text-align:center;font-size:.6rem;font-weight:bold;padding:.1rem;margin-right:.5rem}
.ep-GET{color:#4a9e5c}.ep-POST{color:#d4a843}.ep-PUT{color:#4a7ec9}.ep-DELETE{color:#c94444}
.ep-path{color:var(--cream);flex:1}
.ep-status{font-size:.6rem;color:var(--cm);margin-left:.5rem}
.ep-toggle{margin-left:.5rem}
.btn{font-size:.6rem;padding:.25rem .6rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:var(--bg)}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.6);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:500px;max-width:90vw;max-height:90vh;overflow-y:auto}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust)}
.fr{margin-bottom:.5rem}.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.15rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.35rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:.8rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1>MIRAGE</h1><div style="display:flex;gap:.4rem"><button class="btn btn-p" onclick="openEp()">+ Endpoint</button><button class="btn" onclick="openSvc()">+ Service</button></div></div>
<div class="main" id="main"></div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)cm()"><div class="modal" id="mdl"></div></div>
<script>
const A='/api';let services=[],endpoints=[];
async function load(){const[s,e]=await Promise.all([fetch(A+'/services').then(r=>r.json()),fetch(A+'/endpoints').then(r=>r.json())]);
services=s.services||[];endpoints=e.endpoints||[];render();}
function render(){if(!services.length&&!endpoints.length){document.getElementById('main').innerHTML='<div class="empty">No mock services. Create a service and add endpoints.</div>';return;}
let h='';services.forEach(s=>{
const eps=endpoints.filter(e=>e.service_id===s.id);
h+='<div class="svc"><div class="svc-hdr"><div><span class="svc-name">'+esc(s.name)+'</span>';if(s.prefix)h+=' <span class="svc-prefix">'+esc(s.prefix)+'</span>';h+='</div><button class="btn" onclick="delSvc(\''+s.id+'\')" style="font-size:.5rem;color:var(--cm)">✕</button></div>';
if(!eps.length)h+='<div class="empty" style="padding:1rem">No endpoints</div>';
eps.forEach(e=>{h+='<div class="ep"><span class="ep-method ep-'+e.method+'">'+e.method+'</span><span class="ep-path">'+esc(e.path)+'</span><span class="ep-status">→ '+e.status_code;if(e.delay_ms)h+=' ('+e.delay_ms+'ms delay)';h+='</span><button class="btn" onclick="delEp(\''+e.id+'\')" style="font-size:.5rem;color:var(--cm);margin-left:.3rem">✕</button></div>';});
h+='</div>';});
// Unattached endpoints
const orphans=endpoints.filter(e=>!services.find(s=>s.id===e.service_id));
if(orphans.length){h+='<div class="svc"><div class="svc-hdr"><span class="svc-name">(unattached)</span></div>';orphans.forEach(e=>{h+='<div class="ep"><span class="ep-method ep-'+e.method+'">'+e.method+'</span><span class="ep-path">'+esc(e.path)+'</span><span class="ep-status">→ '+e.status_code+'</span></div>';});h+='</div>';}
document.getElementById('main').innerHTML=h;}
async function delSvc(id){if(confirm('Delete service and all endpoints?')){await fetch(A+'/services/'+id,{method:'DELETE'});load();}}
async function delEp(id){await fetch(A+'/endpoints/'+id,{method:'DELETE'});load();}
function openSvc(){document.getElementById('mdl').innerHTML='<h2>New Mock Service</h2><div class="fr"><label>Name</label><input id="f-n" placeholder="e.g. Payment API"></div><div class="fr"><label>URL Prefix</label><input id="f-p" placeholder="/api/payments"></div><div class="fr"><label>Description</label><input id="f-d"></div><div class="acts"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-p" onclick="subSvc()">Create</button></div>';document.getElementById('mbg').classList.add('open');}
async function subSvc(){await fetch(A+'/services',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:document.getElementById('f-n').value,prefix:document.getElementById('f-p').value,description:document.getElementById('f-d').value})});cm();load();}
function openEp(){let opts=services.map(s=>'<option value="'+s.id+'">'+esc(s.name)+'</option>').join('');
document.getElementById('mdl').innerHTML='<h2>New Mock Endpoint</h2><div class="fr"><label>Service</label><select id="f-s">'+opts+'</select></div><div class="fr"><label>Method</label><select id="f-m"><option>GET</option><option>POST</option><option>PUT</option><option>PATCH</option><option>DELETE</option></select></div><div class="fr"><label>Path</label><input id="f-path" placeholder="/users/{id}"></div><div class="fr"><label>Response Status</label><input id="f-st" type="number" value="200"></div><div class="fr"><label>Response Body</label><textarea id="f-body" rows="6" placeholder=\'{"id":1,"name":"Mock User"}\'></textarea></div><div class="fr"><label>Delay (ms)</label><input id="f-delay" type="number" value="0"></div><div class="acts"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-p" onclick="subEp()">Create</button></div>';document.getElementById('mbg').classList.add('open');}
async function subEp(){await fetch(A+'/endpoints',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({service_id:document.getElementById('f-s').value,method:document.getElementById('f-m').value,path:document.getElementById('f-path').value,status_code:parseInt(document.getElementById('f-st').value),response_body:document.getElementById('f-body').value,delay_ms:parseInt(document.getElementById('f-delay').value)||0})});cm();load();}
function cm(){document.getElementById('mbg').classList.remove('open');}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
