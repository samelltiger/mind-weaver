package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
}

func InitDB(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

func createTables(db *sql.DB) error {
	// Projects table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			language TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_opened_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Sessions table
	db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER,
			name TEXT NOT NULL,
			mode TEXT NOT NULL DEFAULT 'auto',
			exclude_patterns TEXT,
			include_patterns TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			context TEXT,
			FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Messages table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Code contexts table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS code_contexts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER,
			file_path TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
		)
	`)
	return err
}
