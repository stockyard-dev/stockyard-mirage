package server

import (
	"encoding/json"
	"github.com/stockyard-dev/stockyard-mirage/internal/store"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Server struct {
	db      *store.DB
	mux     *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.mux.HandleFunc("GET /api/services", s.listServices)
	s.mux.HandleFunc("POST /api/services", s.createService)
	s.mux.HandleFunc("GET /api/services/{id}", s.getService)
	s.mux.HandleFunc("DELETE /api/services/{id}", s.deleteService)
	s.mux.HandleFunc("GET /api/services/{id}/endpoints", s.listEndpoints)
	s.mux.HandleFunc("POST /api/endpoints", s.createEndpoint)
	s.mux.HandleFunc("DELETE /api/endpoints/{id}", s.deleteEndpoint)
	s.mux.HandleFunc("GET /api/endpoints/{id}/requests", s.listRequests)
	s.mux.HandleFunc("GET /api/requests/recent", s.recentRequests)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/mirage/"})
	})
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)
	return s
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/mock/") {
		s.handleMock(w, r)
		return
	}
	s.mux.ServeHTTP(w, r)
}
func wj(w http.ResponseWriter, c int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	json.NewEncoder(w).Encode(v)
}
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", 302)
}
func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"services": oe(s.db.ListServices())})
}
func (s *Server) createService(w http.ResponseWriter, r *http.Request) {
	var svc store.MockService
	json.NewDecoder(r.Body).Decode(&svc)
	if svc.Name == "" {
		we(w, 400, "name required")
		return
	}
	s.db.CreateService(&svc)
	wj(w, 201, s.db.GetService(svc.ID))
}
func (s *Server) getService(w http.ResponseWriter, r *http.Request) {
	svc := s.db.GetService(r.PathValue("id"))
	if svc == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, svc)
}
func (s *Server) deleteService(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteService(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}
func (s *Server) listEndpoints(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"endpoints": oe(s.db.ListEndpoints(r.PathValue("id")))})
}
func (s *Server) createEndpoint(w http.ResponseWriter, r *http.Request) {
	var ep store.Endpoint
	json.NewDecoder(r.Body).Decode(&ep)
	if ep.Path == "" {
		we(w, 400, "path required")
		return
	}
	s.db.CreateEndpoint(&ep)
	wj(w, 201, s.db.GetEndpoint(ep.ID))
}
func (s *Server) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteEndpoint(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}
func (s *Server) listRequests(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"requests": oe(s.db.ListRequests(r.PathValue("id"), 50))})
}
func (s *Server) recentRequests(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"requests": oe(s.db.RecentRequests(50))})
}
func (s *Server) stats(w http.ResponseWriter, r *http.Request) { wj(w, 200, s.db.Stats()) }
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats()
	wj(w, 200, map[string]any{"status": "ok", "service": "mirage", "endpoints": st.Endpoints, "requests": st.Requests})
}

func (s *Server) handleMock(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/mock")
	ep := s.db.MatchEndpoint(r.Method, path)
	bodyBytes, _ := io.ReadAll(r.Body)
	if ep == nil {
		s.db.LogRequest(&store.RequestLog{Method: r.Method, Path: path, Body: string(bodyBytes), IP: r.RemoteAddr, StatusSent: 404})
		we(w, 404, "no matching mock endpoint")
		return
	}
	if ep.DelayMs > 0 {
		time.Sleep(time.Duration(ep.DelayMs) * time.Millisecond)
	}
	for k, v := range ep.ResponseHeaders {
		w.Header().Set(k, v)
	}
	if ct := w.Header().Get("Content-Type"); ct == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	s.db.LogRequest(&store.RequestLog{EndpointID: ep.ID, Method: r.Method, Path: path, Body: string(bodyBytes), IP: r.RemoteAddr, StatusSent: ep.StatusCode})
	w.WriteHeader(ep.StatusCode)
	w.Write([]byte(ep.ResponseBody))
}

func oe[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

// ─── personalization (auto-added) ──────────────────────────────────

func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("%s: warning: could not parse config.json: %v", "mirage", err)
		return
	}
	s.pCfg = cfg
	log.Printf("%s: loaded personalization from %s", "mirage", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"read body"}`, 400)
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		http.Error(w, `{"error":"invalid json"}`, 400)
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		http.Error(w, `{"error":"save failed"}`, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":"saved"}`))
}
