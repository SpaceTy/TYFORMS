package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"tyforms/internal/models"
)

// ErrAdminExists is returned when creating an admin with a taken username
var ErrAdminExists = errors.New("admin username already exists")

// CreateAdmin inserts a new admin account
func (s *SQLiteStore) CreateAdmin(username, passwordHash string, canManageAdmins bool) (*models.Admin, error) {
	now := time.Now().UTC()
	result, err := s.db.Exec(
		`INSERT INTO admins (username, password_hash, created_at, is_active, can_manage_admins) VALUES (?, ?, ?, TRUE, ?)`,
		username, passwordHash, now, canManageAdmins,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if isUniqueViolation(err) {
			return nil, ErrAdminExists
		}
		return nil, fmt.Errorf("error creating admin: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting admin id: %w", err)
	}

	return &models.Admin{
		ID:              int(id),
		Username:        username,
		CreatedAt:       now,
		IsActive:        true,
		CanManageAdmins: canManageAdmins,
	}, nil
}

// GetAdminByUsername retrieves an active lookup of an admin by username
func (s *SQLiteStore) GetAdminByUsername(username string) (*models.AdminWithHash, error) {
	admin := &models.AdminWithHash{}
	err := s.db.QueryRow(
		`SELECT id, username, password_hash, created_at, is_active, can_manage_admins FROM admins WHERE username = ?`,
		username,
	).Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &admin.CreatedAt, &admin.IsActive, &admin.CanManageAdmins)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting admin by username: %w", err)
	}
	return admin, nil
}

// GetAdminByID retrieves an admin by ID
func (s *SQLiteStore) GetAdminByID(id int) (*models.Admin, error) {
	admin := &models.Admin{}
	err := s.db.QueryRow(
		`SELECT id, username, created_at, is_active, can_manage_admins FROM admins WHERE id = ?`,
		id,
	).Scan(&admin.ID, &admin.Username, &admin.CreatedAt, &admin.IsActive, &admin.CanManageAdmins)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting admin by id: %w", err)
	}
	return admin, nil
}

// ListAdmins retrieves all admin accounts
func (s *SQLiteStore) ListAdmins() ([]*models.Admin, error) {
	rows, err := s.db.Query(`SELECT id, username, created_at, is_active, can_manage_admins FROM admins ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("error listing admins: %w", err)
	}
	defer rows.Close()

	var admins []*models.Admin
	for rows.Next() {
		admin := &models.Admin{}
		if err := rows.Scan(&admin.ID, &admin.Username, &admin.CreatedAt, &admin.IsActive, &admin.CanManageAdmins); err != nil {
			return nil, fmt.Errorf("error scanning admin: %w", err)
		}
		admins = append(admins, admin)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating admins: %w", err)
	}
	return admins, nil
}

// CountAdmins returns the number of admin accounts
func (s *SQLiteStore) CountAdmins() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM admins`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting admins: %w", err)
	}
	return count, nil
}

// SetAdminPassword updates an admin's password hash
func (s *SQLiteStore) SetAdminPassword(id int, passwordHash string) error {
	_, err := s.db.Exec(`UPDATE admins SET password_hash = ? WHERE id = ?`, passwordHash, id)
	if err != nil {
		return fmt.Errorf("error updating admin password: %w", err)
	}
	return nil
}

// SetAdminPermissions updates whether an admin may access and manage the
// admins page
func (s *SQLiteStore) SetAdminPermissions(id int, canManageAdmins bool) error {
	_, err := s.db.Exec(`UPDATE admins SET can_manage_admins = ? WHERE id = ?`, canManageAdmins, id)
	if err != nil {
		return fmt.Errorf("error updating admin permissions: %w", err)
	}
	return nil
}

// DeleteAdmin removes an admin account and all of its sessions
func (s *SQLiteStore) DeleteAdmin(id int) error {
	if _, err := s.db.Exec(`DELETE FROM admin_sessions WHERE admin_id = ?`, id); err != nil {
		return fmt.Errorf("error deleting admin sessions: %w", err)
	}
	if _, err := s.db.Exec(`DELETE FROM admins WHERE id = ?`, id); err != nil {
		return fmt.Errorf("error deleting admin: %w", err)
	}
	return nil
}

// CreateSession stores a session token of the given kind for an admin
func (s *SQLiteStore) CreateSession(token string, adminID int, kind string, expiresAt time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO admin_sessions (token, admin_id, kind, created_at, expires_at) VALUES (?, ?, ?, ?, ?)`,
		token, adminID, kind, time.Now().UTC(), expiresAt,
	)
	if err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}
	return nil
}

// GetValidSession resolves a token of the given kind to its admin ID.
// Expired tokens are removed on sight.
func (s *SQLiteStore) GetValidSession(token, kind string) (int, bool, error) {
	var adminID int
	var expiresAt time.Time
	err := s.db.QueryRow(
		`SELECT admin_id, expires_at FROM admin_sessions WHERE token = ? AND kind = ?`,
		token, kind,
	).Scan(&adminID, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("error getting session: %w", err)
	}

	if time.Now().UTC().After(expiresAt) {
		if err := s.DeleteSession(token); err != nil {
			return 0, false, err
		}
		return 0, false, nil
	}
	return adminID, true, nil
}

// DeleteSession revokes a single session token
func (s *SQLiteStore) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM admin_sessions WHERE token = ?`, token)
	if err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes expired session tokens
func (s *SQLiteStore) DeleteExpiredSessions() error {
	_, err := s.db.Exec(`DELETE FROM admin_sessions WHERE expires_at < ?`, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("error deleting expired sessions: %w", err)
	}
	return nil
}

// isUniqueViolation checks for SQLite unique constraint errors
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
