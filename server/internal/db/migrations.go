package db

import (
	"database/sql"
	"fmt"
)

func Migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id          TEXT PRIMARY KEY,
			subdomain   TEXT UNIQUE NOT NULL,
			auth_token  TEXT NOT NULL,
			local_port  INTEGER NOT NULL,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_seen   DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS access_tokens (
			token       TEXT PRIMARY KEY,
			session_id  TEXT NOT NULL,
			label       TEXT,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS request_logs (
			id          TEXT PRIMARY KEY,
			session_id  TEXT NOT NULL,
			method      TEXT NOT NULL,
			path        TEXT NOT NULL,
			status      INTEGER,
			duration_ms INTEGER,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS feedbacks (
			id          TEXT PRIMARY KEY,
			session_id  TEXT NOT NULL,
			page_url    TEXT NOT NULL,
			element_css TEXT,
			x_percent   REAL,
			y_percent   REAL,
			comment     TEXT NOT NULL,
			author_name TEXT,
			resolved    INTEGER DEFAULT 0,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}
