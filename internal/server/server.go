package server

import (
	"encoding/json"
	"net/http"

	"github.com/stockyard-dev/stockyard-mirage/internal/store"
)

type Server struct {
	db     *store.DB
	limits Limits
	mux    *http.ServeMux
}

func New(db *store.DB, tier string) *Server {
	s := &Server{
		db:     db,
		limits: LimitsFor(tier),
		mux:    http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{Addr: addr, Handler: s.mux}
	return srv.ListenAndServe()
}

func (s *Server) routes() {
	// Admin API
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/version", s.handleVersion)
	s.mux.HandleFunc("GET /api/limits", s.handleLimits)
	s.mux.HandleFunc("GET /api/endpoints", s.handleListEndpoints)
	s.mux.HandleFunc("POST /api/endpoints", s.handleCreateEndpoint)
	s.mux.HandleFunc("GET /api/endpoints/{id}", s.handleGetEndpoint)
	s.mux.HandleFunc("PUT /api/endpoints/{id}", s.handleUpdateEndpoint)
	s.mux.HandleFunc("DELETE /api/endpoints/{id}", s.handleDeleteEndpoint)
	s.mux.HandleFunc("GET /api/logs", s.handleListLogs)
	s.mux.HandleFunc("GET /api/stats", s.handleStats)
	// Dashboard
	s.mux.HandleFunc("GET /", s.handleUI)
	// Mock catch-all — must be last
	s.mux.HandleFunc("/mock/", s.handleMock)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "stockyard-mirage"})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": "0.1.0", "service": "stockyard-mirage"})
}

func (s *Server) handleLimits(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tier":        s.limits.Tier,
		"description": s.limits.Description,
		"is_pro":      s.limits.IsPro(),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	endpoints, _ := s.db.CountEndpoints()
	logs, _ := s.db.CountLogs()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"endpoints": endpoints,
		"logs":      logs,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
