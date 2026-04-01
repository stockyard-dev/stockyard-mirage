package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
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
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS endpoints (
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
     );`)
	return err
}
