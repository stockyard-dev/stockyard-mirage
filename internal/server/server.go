package server
import ("encoding/json";"io";"log";"net/http";"strings";"time";"github.com/stockyard-dev/stockyard-mirage/internal/store")
type Server struct{db *store.DB;mux *http.ServeMux}
func New(db *store.DB)*Server{s:=&Server{db:db,mux:http.NewServeMux()}
s.mux.HandleFunc("GET /api/services",s.listServices);s.mux.HandleFunc("POST /api/services",s.createService);s.mux.HandleFunc("GET /api/services/{id}",s.getService);s.mux.HandleFunc("DELETE /api/services/{id}",s.deleteService)
s.mux.HandleFunc("GET /api/services/{id}/endpoints",s.listEndpoints);s.mux.HandleFunc("POST /api/endpoints",s.createEndpoint);s.mux.HandleFunc("DELETE /api/endpoints/{id}",s.deleteEndpoint)
s.mux.HandleFunc("GET /api/endpoints/{id}/requests",s.listRequests);s.mux.HandleFunc("GET /api/requests/recent",s.recentRequests)
s.mux.HandleFunc("GET /api/stats",s.stats);s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root);return s}
func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){
if strings.HasPrefix(r.URL.Path,"/mock/"){s.handleMock(w,r);return};s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)listServices(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"services":oe(s.db.ListServices())})}
func(s *Server)createService(w http.ResponseWriter,r *http.Request){var svc store.MockService;json.NewDecoder(r.Body).Decode(&svc);if svc.Name==""{we(w,400,"name required");return};s.db.CreateService(&svc);wj(w,201,s.db.GetService(svc.ID))}
func(s *Server)getService(w http.ResponseWriter,r *http.Request){svc:=s.db.GetService(r.PathValue("id"));if svc==nil{we(w,404,"not found");return};wj(w,200,svc)}
func(s *Server)deleteService(w http.ResponseWriter,r *http.Request){s.db.DeleteService(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)listEndpoints(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"endpoints":oe(s.db.ListEndpoints(r.PathValue("id")))})}
func(s *Server)createEndpoint(w http.ResponseWriter,r *http.Request){var ep store.Endpoint;json.NewDecoder(r.Body).Decode(&ep);if ep.Path==""{we(w,400,"path required");return};s.db.CreateEndpoint(&ep);wj(w,201,s.db.GetEndpoint(ep.ID))}
func(s *Server)deleteEndpoint(w http.ResponseWriter,r *http.Request){s.db.DeleteEndpoint(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)listRequests(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"requests":oe(s.db.ListRequests(r.PathValue("id"),50))})}
func(s *Server)recentRequests(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"requests":oe(s.db.RecentRequests(50))})}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){st:=s.db.Stats();wj(w,200,map[string]any{"status":"ok","service":"mirage","endpoints":st.Endpoints,"requests":st.Requests})}

func(s *Server)handleMock(w http.ResponseWriter,r *http.Request){
path:=strings.TrimPrefix(r.URL.Path,"/mock")
ep:=s.db.MatchEndpoint(r.Method,path)
bodyBytes,_:=io.ReadAll(r.Body)
if ep==nil{
s.db.LogRequest(&store.RequestLog{Method:r.Method,Path:path,Body:string(bodyBytes),IP:r.RemoteAddr,StatusSent:404})
we(w,404,"no matching mock endpoint");return}
if ep.DelayMs>0{time.Sleep(time.Duration(ep.DelayMs)*time.Millisecond)}
for k,v:=range ep.ResponseHeaders{w.Header().Set(k,v)}
if ct:=w.Header().Get("Content-Type");ct==""{w.Header().Set("Content-Type","application/json")}
s.db.LogRequest(&store.RequestLog{EndpointID:ep.ID,Method:r.Method,Path:path,Body:string(bodyBytes),IP:r.RemoteAddr,StatusSent:ep.StatusCode})
w.WriteHeader(ep.StatusCode);w.Write([]byte(ep.ResponseBody))}

func oe[T any](s []T)[]T{if s==nil{return[]T{}};return s}
func init(){log.SetFlags(log.LstdFlags|log.Lshortfile)}
