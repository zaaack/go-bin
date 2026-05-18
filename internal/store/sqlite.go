package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS shares (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			kind TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			content_text TEXT NOT NULL DEFAULT '',
			stored_path TEXT NOT NULL DEFAULT '',
			original_name TEXT NOT NULL DEFAULT '',
			mime_type TEXT NOT NULL DEFAULT '',
			size_bytes INTEGER NOT NULL DEFAULT 0,
			is_public INTEGER NOT NULL DEFAULT 1,
			is_pinned INTEGER NOT NULL DEFAULT 0,
			expires_at DATETIME NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_shares_public_pin_created
		ON shares(is_public, is_pinned, created_at DESC);

		CREATE INDEX IF NOT EXISTS idx_shares_expires_at
		ON shares(expires_at);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("init sqlite schema: %w", err)
	}

	return db, nil
}
