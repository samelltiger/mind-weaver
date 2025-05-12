package db

import (
	"time"
)

type Project struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Language     string    `json:"language"`
	CreatedAt    time.Time `json:"created_at"`
	LastOpenedAt time.Time `json:"last_opened_at"`
}

type Session struct {
	ID              int64     `json:"id"`
	ProjectID       int64     `json:"project_id"`
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Context         string    `json:"context"` // JSON string of context info
	Mode            string    `json:"mode"`
	ExcludePatterns string    `json:"exclude_patterns,omitempty"`
	IncludePatterns string    `json:"include_patterns,omitempty"`
}

type FileInfo struct {
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}

type Message struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	Role      string    `json:"role"` // "user" or "ai"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type CodeContext struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	FilePath  string    `json:"file_path"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
