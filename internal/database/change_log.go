package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"tyforms/internal/models"
)

// RecordChange appends a new node to an application's change tree. The parent
// is the change that was current when this change was made, so concurrent
// edits from the same base state become sibling branches. The first change for
// an application becomes the tree root.
func (s *SQLiteStore) RecordChange(applicationID int, adminID *int, adminUsername, action string, changes []models.FieldChange) error {
	if changes == nil {
		changes = []models.FieldChange{}
	}
	changesJSON, err := json.Marshal(changes)
	if err != nil {
		return fmt.Errorf("error marshaling changes: %w", err)
	}

	var parentID, rootID *int
	var parentRoot int
	err = s.db.QueryRow(
		`SELECT id, COALESCE(root_id, id) FROM change_log WHERE application_id = ? ORDER BY id DESC LIMIT 1`,
		applicationID,
	).Scan(&parentID, &parentRoot)
	switch {
	case err == nil:
		rootID = &parentRoot
	case errors.Is(err, sql.ErrNoRows):
		// First change for this application; it becomes the root.
	default:
		return fmt.Errorf("error finding parent change: %w", err)
	}

	result, err := s.db.Exec(
		`INSERT INTO change_log (application_id, parent_id, root_id, admin_id, admin_username, action, changes, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		applicationID, parentID, rootID, adminID, adminUsername, action, string(changesJSON), time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("error recording change: %w", err)
	}

	// Self-root the first change of the tree.
	if rootID == nil {
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("error getting change id: %w", err)
		}
		if _, err := s.db.Exec(`UPDATE change_log SET root_id = id WHERE id = ? AND root_id IS NULL`, id); err != nil {
			return fmt.Errorf("error linking change root: %w", err)
		}
	}

	return nil
}

// GetApplicationChanges returns the full change tree (in insertion order) for one application
func (s *SQLiteStore) GetApplicationChanges(applicationID int) ([]*models.ChangeEntry, error) {
	return s.queryChanges(`SELECT id, application_id, parent_id, root_id, admin_id, admin_username, action, changes, created_at FROM change_log WHERE application_id = ? ORDER BY id ASC`, applicationID)
}

// GetRecentChanges returns the latest changes across all applications (newest first)
func (s *SQLiteStore) GetRecentChanges(limit int) ([]*models.ChangeEntry, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	return s.queryChanges(fmt.Sprintf(`SELECT id, application_id, parent_id, root_id, admin_id, admin_username, action, changes, created_at FROM change_log ORDER BY id DESC LIMIT %d`, limit))
}

func (s *SQLiteStore) queryChanges(query string, args ...any) ([]*models.ChangeEntry, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying changes: %w", err)
	}
	defer rows.Close()

	var entries []*models.ChangeEntry
	for rows.Next() {
		entry := &models.ChangeEntry{}
		var changesJSON string
		err := rows.Scan(
			&entry.ID,
			&entry.ApplicationID,
			&entry.ParentID,
			&entry.RootID,
			&entry.AdminID,
			&entry.AdminUsername,
			&entry.Action,
			&changesJSON,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning change: %w", err)
		}
		if err := json.Unmarshal([]byte(changesJSON), &entry.Changes); err != nil {
			return nil, fmt.Errorf("error unmarshaling changes: %w", err)
		}
		if entry.Changes == nil {
			entry.Changes = []models.FieldChange{}
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating changes: %w", err)
	}

	if entries == nil {
		entries = []*models.ChangeEntry{}
	}
	return entries, nil
}
