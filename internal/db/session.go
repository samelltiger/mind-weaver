package db

import (
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Session CRUD operations
func (db *Database) CreateSession(projectID int64, name string, mode string, excludePatterns string, includePatterns string, contextInfo string) (int64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO sessions (project_id, name, include_patterns, mode, exclude_patterns, context) VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(projectID, name, includePatterns, mode, excludePatterns, contextInfo)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (db *Database) GetSession(id int64) (*Session, error) {
	session := &Session{}
	err := db.QueryRow(`
		SELECT id, project_id, name, created_at, updated_at, context, include_patterns, mode, exclude_patterns
		FROM sessions WHERE id = ?
	`, id).Scan(
		&session.ID, &session.ProjectID, &session.Name,
		&session.CreatedAt, &session.UpdatedAt, &session.Context,
		&session.IncludePatterns, &session.Mode, &session.ExcludePatterns,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (db *Database) ListProjectSessions(projectID int64) ([]*Session, error) {
	rows, err := db.Query(`
		SELECT id, project_id, name, created_at, updated_at, context, include_patterns, mode, exclude_patterns
		FROM sessions WHERE project_id = ? ORDER BY updated_at DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []*Session{}
	for rows.Next() {
		session := &Session{}
		err := rows.Scan(
			&session.ID, &session.ProjectID, &session.Name,
			&session.CreatedAt, &session.UpdatedAt, &session.Context,
			&session.IncludePatterns, &session.Mode, &session.ExcludePatterns,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// Message CRUD operations
func (db *Database) AddMessage(sessionID int64, role, content string) (int64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO messages (session_id, role, content) VALUES (?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Update session updated_at timestamp
	_, err = db.Exec(`UPDATE sessions SET updated_at = ? WHERE id = ?`, time.Now(), sessionID)
	if err != nil {
		log.Printf("Failed to update session timestamp: %v", err)
	}

	res, err := stmt.Exec(sessionID, role, content)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (db *Database) GetSessionMessages(sessionID int64) ([]*Message, error) {
	rows, err := db.Query(`
		SELECT id, session_id, role, content, timestamp
		FROM messages WHERE session_id = ? ORDER BY timestamp
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*Message{}
	for rows.Next() {
		message := &Message{}
		err := rows.Scan(
			&message.ID, &message.SessionID, &message.Role,
			&message.Content, &message.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

// Code Context operations
func (db *Database) SaveCodeContext(sessionID int64, filePath, content string) (int64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO code_contexts (session_id, file_path, content) VALUES (?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(sessionID, filePath, content)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (db *Database) GetSessionContexts(sessionID int64) ([]*CodeContext, error) {
	rows, err := db.Query(`
		SELECT id, session_id, file_path, content, created_at
		FROM code_contexts WHERE session_id = ?
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contexts := []*CodeContext{}
	for rows.Next() {
		ctx := &CodeContext{}
		err := rows.Scan(
			&ctx.ID, &ctx.SessionID, &ctx.FilePath,
			&ctx.Content, &ctx.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		contexts = append(contexts, ctx)
	}
	return contexts, nil
}

// UpdateSession updates a session's details
func (db *Database) UpdateSession(
	id int64,
	name string,
	mode string,
	excludePatterns string,
	includePatterns string,
) error {
	updateFields := []string{}
	args := []interface{}{}

	// Only update fields that are provided
	if name != "" {
		updateFields = append(updateFields, "name = ?")
		args = append(args, name)
	}

	if mode != "" {
		updateFields = append(updateFields, "mode = ?")
		args = append(args, mode)
	}

	if excludePatterns != "" {
		updateFields = append(updateFields, "exclude_patterns = ?")
		args = append(args, excludePatterns)
	}

	if includePatterns != "" {
		updateFields = append(updateFields, "include_patterns = ?")
		args = append(args, includePatterns)
	}

	// Add updated_at field
	updateFields = append(updateFields, "updated_at = ?")
	args = append(args, time.Now())

	// Add session ID to args
	args = append(args, id)

	// If no fields to update, return
	if len(updateFields) == 1 { // Only has updated_at
		return nil
	}

	// Build and execute query
	query := "UPDATE sessions SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
	_, err := db.Exec(query, args...)
	return err
}

// DeleteSession removes a session and all related data
func (db *Database) DeleteSession(id int64) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Delete messages first (foreign key constraint)
	_, err = tx.Exec("DELETE FROM messages WHERE session_id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete code contexts
	_, err = tx.Exec("DELETE FROM code_contexts WHERE session_id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete the session
	_, err = tx.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// DeleteMessage 删除指定的消息记录
func (db *Database) DeleteMessage(id int64) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Delete messages
	_, err = tx.Exec("DELETE FROM messages WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// DeleteAllMessage 删除指定回话下的所有消息记录
func (db *Database) DeleteAllMessage(id int64) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Delete messages
	_, err = tx.Exec("DELETE FROM messages WHERE session_id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
