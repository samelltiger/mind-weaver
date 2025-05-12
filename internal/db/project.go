package db

import (
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Project CRUD operations
func (db *Database) CreateProject(name, path, language string) (int64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO projects (name, path, language) VALUES (?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(name, path, language)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (db *Database) GetProject(id int64) (*Project, error) {
	project := &Project{}
	err := db.QueryRow(`
		SELECT id, name, path, language, created_at, last_opened_at 
		FROM projects WHERE id = ?
	`, id).Scan(
		&project.ID, &project.Name, &project.Path, &project.Language,
		&project.CreatedAt, &project.LastOpenedAt,
	)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (db *Database) GetProjectByPath(path string) (*Project, error) {
	project := &Project{}
	err := db.QueryRow(`
		SELECT id, name, path, language, created_at, last_opened_at 
		FROM projects WHERE path = ?
	`, path).Scan(
		&project.ID, &project.Name, &project.Path, &project.Language,
		&project.CreatedAt, &project.LastOpenedAt,
	)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (db *Database) UpdateProjectLastOpened(id int64) error {
	_, err := db.Exec(`
		UPDATE projects SET last_opened_at = ? WHERE id = ?
	`, time.Now(), id)
	return err
}

func (db *Database) ListProjects() ([]*Project, error) {
	rows, err := db.Query(`
		SELECT id, name, path, language, created_at, last_opened_at 
		FROM projects ORDER BY last_opened_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*Project{}
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Path, &project.Language,
			&project.CreatedAt, &project.LastOpenedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}

// UpdateProject updates project information
func (db *Database) UpdateProject(id int64, name, path, language string) error {
	_, err := db.Exec(`
        UPDATE projects 
        SET name = ?, path = ?, language = ?
        WHERE id = ?
    `, name, path, language, id)
	return err
}
