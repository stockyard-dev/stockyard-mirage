package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

type Endpoint struct {
	ID          int64     `json:"id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	StatusCode  int       `json:"status_code"`
	ResponseBody string   `json:"response_body"`
	ContentType string    `json:"content_type"`
	DelayMs     int       `json:"delay_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

type RequestLog struct {
	ID             int64     `json:"id"`
	EndpointID     *int64    `json:"endpoint_id"`
	Method         string    `json:"method"`
	Path           string    `json:"path"`
	RequestHeaders string    `json:"request_headers"`
	RequestBody    string    `json:"request_body"`
	RespondedAt    time.Time `json:"responded_at"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}
	dsn := filepath.Join(dataDir, "mirage.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS endpoints (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			method TEXT NOT NULL DEFAULT 'GET',
			path TEXT NOT NULL,
			status_code INTEGER NOT NULL DEFAULT 200,
			response_body TEXT NOT NULL DEFAULT '{}',
			content_type TEXT NOT NULL DEFAULT 'application/json',
			delay_ms INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS request_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			endpoint_id INTEGER,
			method TEXT,
			path TEXT,
			request_headers TEXT,
			request_body TEXT,
			responded_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func (db *DB) ListEndpoints() ([]Endpoint, error) {
	rows, err := db.Query(`SELECT id, method, path, status_code, response_body, content_type, delay_ms, created_at FROM endpoints ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Endpoint
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(&e.ID, &e.Method, &e.Path, &e.StatusCode, &e.ResponseBody, &e.ContentType, &e.DelayMs, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (db *DB) CreateEndpoint(e *Endpoint) error {
	res, err := db.Exec(
		`INSERT INTO endpoints (method, path, status_code, response_body, content_type, delay_ms) VALUES (?, ?, ?, ?, ?, ?)`,
		e.Method, e.Path, e.StatusCode, e.ResponseBody, e.ContentType, e.DelayMs,
	)
	if err != nil {
		return err
	}
	e.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) GetEndpoint(id int64) (*Endpoint, error) {
	e := &Endpoint{}
	err := db.QueryRow(`SELECT id, method, path, status_code, response_body, content_type, delay_ms, created_at FROM endpoints WHERE id = ?`, id).
		Scan(&e.ID, &e.Method, &e.Path, &e.StatusCode, &e.ResponseBody, &e.ContentType, &e.DelayMs, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return e, err
}

func (db *DB) UpdateEndpoint(e *Endpoint) error {
	_, err := db.Exec(
		`UPDATE endpoints SET method=?, path=?, status_code=?, response_body=?, content_type=?, delay_ms=? WHERE id=?`,
		e.Method, e.Path, e.StatusCode, e.ResponseBody, e.ContentType, e.DelayMs, e.ID,
	)
	return err
}

func (db *DB) DeleteEndpoint(id int64) error {
	_, err := db.Exec(`DELETE FROM endpoints WHERE id = ?`, id)
	return err
}

func (db *DB) FindEndpointByMethodPath(method, path string) (*Endpoint, error) {
	e := &Endpoint{}
	err := db.QueryRow(`SELECT id, method, path, status_code, response_body, content_type, delay_ms, created_at FROM endpoints WHERE method=? AND path=?`, method, path).
		Scan(&e.ID, &e.Method, &e.Path, &e.StatusCode, &e.ResponseBody, &e.ContentType, &e.DelayMs, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return e, err
}

func (db *DB) LogRequest(log *RequestLog) error {
	_, err := db.Exec(
		`INSERT INTO request_log (endpoint_id, method, path, request_headers, request_body) VALUES (?, ?, ?, ?, ?)`,
		log.EndpointID, log.Method, log.Path, log.RequestHeaders, log.RequestBody,
	)
	return err
}

func (db *DB) ListRequestLogs(limit int) ([]RequestLog, error) {
	rows, err := db.Query(`SELECT id, endpoint_id, method, path, request_headers, request_body, responded_at FROM request_log ORDER BY responded_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RequestLog
	for rows.Next() {
		var l RequestLog
		if err := rows.Scan(&l.ID, &l.EndpointID, &l.Method, &l.Path, &l.RequestHeaders, &l.RequestBody, &l.RespondedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

func (db *DB) CountEndpoints() (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM endpoints`).Scan(&n)
	return n, err
}

func (db *DB) CountLogs() (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM request_log`).Scan(&n)
	return n, err
}
