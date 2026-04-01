package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stockyard-dev/stockyard-mirage/internal/store"
)

func (s *Server) handleListEndpoints(w http.ResponseWriter, r *http.Request) {
	endpoints, err := s.db.ListEndpoints()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if endpoints == nil {
		endpoints = []store.Endpoint{}
	}
	writeJSON(w, http.StatusOK, endpoints)
}

func (s *Server) handleCreateEndpoint(w http.ResponseWriter, r *http.Request) {
	if !s.limits.IsPro() {
		count, _ := s.db.CountEndpoints()
		if count >= 10 {
			writeError(w, http.StatusForbidden, "free tier limit: 10 endpoints. Upgrade to Pro for unlimited.")
			return
		}
	}
	var e store.Endpoint
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if e.Path == "" {
		writeError(w, http.StatusBadRequest, "path required")
		return
	}
	if e.Method == "" {
		e.Method = "GET"
	}
	if e.StatusCode == 0 {
		e.StatusCode = 200
	}
	if e.ContentType == "" {
		e.ContentType = "application/json"
	}
	if e.ResponseBody == "" {
		e.ResponseBody = "{}"
	}
	if err := s.db.CreateEndpoint(&e); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (s *Server) handleGetEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	e, err := s.db.GetEndpoint(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if e == nil {
		writeError(w, http.StatusNotFound, "endpoint not found")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) handleUpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	existing, err := s.db.GetEndpoint(id)
	if err != nil || existing == nil {
		writeError(w, http.StatusNotFound, "endpoint not found")
		return
	}
	var e store.Endpoint
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	e.ID = id
	if e.Method == "" {
		e.Method = existing.Method
	}
	if e.Path == "" {
		e.Path = existing.Path
	}
	if e.StatusCode == 0 {
		e.StatusCode = existing.StatusCode
	}
	if e.ContentType == "" {
		e.ContentType = existing.ContentType
	}
	if err := s.db.UpdateEndpoint(&e); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) handleDeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.db.DeleteEndpoint(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleListLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if !s.limits.IsPro() && limit > 500 {
		limit = 500
	}
	logs, err := s.db.ListRequestLogs(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if logs == nil {
		logs = []store.RequestLog{}
	}
	writeJSON(w, http.StatusOK, logs)
}

// handleMock serves all /mock/* requests against configured endpoints
func (s *Server) handleMock(w http.ResponseWriter, r *http.Request) {
	// Strip /mock prefix
	mockPath := strings.TrimPrefix(r.URL.Path, "/mock")
	if mockPath == "" {
		mockPath = "/"
	}

	// Read body for logging
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))

	// Build headers string
	var hdrs strings.Builder
	for k, vs := range r.Header {
		fmt.Fprintf(&hdrs, "%s: %s\n", k, strings.Join(vs, ", "))
	}

	// Look up endpoint
	endpoint, err := s.db.FindEndpointByMethodPath(r.Method, mockPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error")
		return
	}

	logEntry := &store.RequestLog{
		Method:         r.Method,
		Path:           mockPath,
		RequestHeaders: hdrs.String(),
		RequestBody:    string(body),
	}

	if endpoint == nil {
		// Also try wildcard match by path only
		endpoint, err = s.db.FindEndpointByMethodPath("*", mockPath)
	}

	if endpoint != nil {
		logEntry.EndpointID = &endpoint.ID
		if endpoint.DelayMs > 0 {
			time.Sleep(time.Duration(endpoint.DelayMs) * time.Millisecond)
		}
		_ = s.db.LogRequest(logEntry)
		w.Header().Set("Content-Type", endpoint.ContentType)
		w.WriteHeader(endpoint.StatusCode)
		w.Write([]byte(endpoint.ResponseBody))
		return
	}

	_ = s.db.LogRequest(logEntry)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{"error":"no mock endpoint configured for %s %s"}`, r.Method, mockPath)
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(dashboardHTML)
}
