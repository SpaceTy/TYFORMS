package database

import (
	"database/sql"
	"fmt"
	"tyforms/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore handles all database operations
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLiteStore instance
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("error creating tables: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// createTables creates the necessary tables in the database
func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS applications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		discord_username TEXT NOT NULL,
		minecraft_username TEXT NOT NULL UNIQUE,
		age INTEGER NOT NULL,
		favorite_about_minecraft TEXT NOT NULL,
		understanding_of_smp TEXT NOT NULL,
		joined_discord BOOLEAN NOT NULL,
		submission_date DATETIME NOT NULL,
		is_reviewed BOOLEAN NOT NULL DEFAULT FALSE,
		reviewed_at DATETIME,
		review_notes TEXT,
		acceptance_status TEXT NOT NULL DEFAULT 'pending'
	)`

	_, err := db.Exec(query)
	return err
}

// CreateApplication inserts a new application into the database
func (s *SQLiteStore) CreateApplication(app *models.MinecraftApplication) error {
	query := `
	INSERT INTO applications (
		discord_username, minecraft_username, age, favorite_about_minecraft,
		understanding_of_smp, joined_discord, submission_date, is_reviewed,
		acceptance_status
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query,
		app.DiscordUsername,
		app.MinecraftUsername,
		app.Age,
		app.FavoriteAboutMinecraft,
		app.UnderstandingOfSMP,
		app.JoinedDiscord,
		app.SubmissionDate,
		app.IsReviewed,
		app.AcceptanceStatus,
	)
	if err != nil {
		return fmt.Errorf("error creating application: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	app.ID = int(id)
	return nil
}

// GetApplication retrieves an application by ID
func (s *SQLiteStore) GetApplication(id int) (*models.MinecraftApplication, error) {
	query := `SELECT * FROM applications WHERE id = ?`
	app := &models.MinecraftApplication{}

	err := s.db.QueryRow(query, id).Scan(
		&app.ID,
		&app.DiscordUsername,
		&app.MinecraftUsername,
		&app.Age,
		&app.FavoriteAboutMinecraft,
		&app.UnderstandingOfSMP,
		&app.JoinedDiscord,
		&app.SubmissionDate,
		&app.IsReviewed,
		&app.ReviewedAt,
		&app.ReviewNotes,
		&app.AcceptanceStatus,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting application: %w", err)
	}

	return app, nil
}

// GetAllApplications retrieves all applications
func (s *SQLiteStore) GetAllApplications() ([]*models.MinecraftApplication, error) {
	query := `SELECT * FROM applications`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying applications: %w", err)
	}
	defer rows.Close()

	var applications []*models.MinecraftApplication
	for rows.Next() {
		app := &models.MinecraftApplication{}
		err := rows.Scan(
			&app.ID,
			&app.DiscordUsername,
			&app.MinecraftUsername,
			&app.Age,
			&app.FavoriteAboutMinecraft,
			&app.UnderstandingOfSMP,
			&app.JoinedDiscord,
			&app.SubmissionDate,
			&app.IsReviewed,
			&app.ReviewedAt,
			&app.ReviewNotes,
			&app.AcceptanceStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning application: %w", err)
		}
		applications = append(applications, app)
	}

	return applications, nil
}

// UpdateApplication updates an existing application
func (s *SQLiteStore) UpdateApplication(app *models.MinecraftApplication) error {
	query := `
	UPDATE applications SET
		discord_username = ?,
		minecraft_username = ?,
		age = ?,
		favorite_about_minecraft = ?,
		understanding_of_smp = ?,
		joined_discord = ?,
		submission_date = ?,
		is_reviewed = ?,
		reviewed_at = ?,
		review_notes = ?,
		acceptance_status = ?
	WHERE id = ?`

	_, err := s.db.Exec(query,
		app.DiscordUsername,
		app.MinecraftUsername,
		app.Age,
		app.FavoriteAboutMinecraft,
		app.UnderstandingOfSMP,
		app.JoinedDiscord,
		app.SubmissionDate,
		app.IsReviewed,
		app.ReviewedAt,
		app.ReviewNotes,
		app.AcceptanceStatus,
		app.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating application: %w", err)
	}

	return nil
}

// DeleteApplication removes an application by ID
func (s *SQLiteStore) DeleteApplication(id int) error {
	query := `DELETE FROM applications WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting application: %w", err)
	}
	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
