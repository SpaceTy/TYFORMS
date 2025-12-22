package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	log.Printf("Preparing to insert application for user: %s", app.MinecraftUsername)

	query := `
	INSERT INTO applications (
		discord_username, minecraft_username, age, favorite_about_minecraft,
		understanding_of_smp, joined_discord, submission_date, is_reviewed,
		acceptance_status
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	log.Printf("Executing database query")
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
		log.Printf("Database error during Exec: %v", err)
		return fmt.Errorf("error creating application: %w", err)
	}

	log.Printf("Getting last insert ID")
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	app.ID = int(id)
	log.Printf("Successfully created application with ID: %d", app.ID)
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

// fuzzyMatch checks if characters in needle appear in order in haystack (case-insensitive)
// This replicates the frontend fuzzy search algorithm
func fuzzyMatch(haystack, needle string) bool {
	if needle == "" {
		return true
	}

	// Convert to lowercase for case-insensitive matching
	h := strings.ToLower(haystack)
	n := strings.ToLower(needle)

	// Convert to runes to handle Unicode properly
	hRunes := []rune(h)
	nRunes := []rune(n)

	hIndex := 0
	for _, nChar := range nRunes {
		found := false
		for hIndex < len(hRunes) {
			if hRunes[hIndex] == nChar {
				found = true
				hIndex++
				break
			}
			hIndex++
		}
		if !found {
			return false
		}
	}

	return true
}

// matchesSearch checks if an application matches the search query in any of the specified fields
func matchesSearch(app *models.MinecraftApplication, query string, fields []string) bool {
	if query == "" {
		return true
	}

	// Create a set of fields to search
	fieldSet := make(map[string]bool)
	for _, field := range fields {
		fieldSet[field] = true
	}

	// If no fields specified, search all fields
	if len(fields) == 0 {
		fieldSet["discordUsername"] = true
		fieldSet["minecraftUsername"] = true
		fieldSet["favoriteAboutMinecraft"] = true
		fieldSet["understandingOfSMP"] = true
		fieldSet["id"] = true
	}

	// Check each field
	if fieldSet["discordUsername"] && fuzzyMatch(app.DiscordUsername, query) {
		return true
	}
	if fieldSet["minecraftUsername"] && fuzzyMatch(app.MinecraftUsername, query) {
		return true
	}
	if fieldSet["favoriteAboutMinecraft"] && fuzzyMatch(app.FavoriteAboutMinecraft, query) {
		return true
	}
	if fieldSet["understandingOfSMP"] && fuzzyMatch(app.UnderstandingOfSMP, query) {
		return true
	}
	if fieldSet["id"] && fuzzyMatch(strconv.Itoa(app.ID), query) {
		return true
	}

	return false
}

// SearchApplications retrieves applications with fuzzy search and pagination
func (s *SQLiteStore) SearchApplications(query string, fields []string, page, pageSize int) ([]*models.MinecraftApplication, int, error) {
	// Validate and set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Fetch all applications ordered by ID DESC
	query_sql := `SELECT * FROM applications ORDER BY id DESC`
	rows, err := s.db.Query(query_sql)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying applications: %w", err)
	}
	defer rows.Close()

	// Scan all applications
	var allApplications []*models.MinecraftApplication
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
			return nil, 0, fmt.Errorf("error scanning application: %w", err)
		}
		allApplications = append(allApplications, app)
	}

	// Apply fuzzy search filter
	var filteredApplications []*models.MinecraftApplication
	for _, app := range allApplications {
		if matchesSearch(app, query, fields) {
			filteredApplications = append(filteredApplications, app)
		}
	}

	// Calculate total count after filtering
	total := len(filteredApplications)

	// Apply pagination
	offset := (page - 1) * pageSize
	end := offset + pageSize

	// Handle edge cases
	if offset >= total {
		// Page is beyond available results, return empty
		return []*models.MinecraftApplication{}, total, nil
	}
	if end > total {
		end = total
	}

	// Return paginated results
	paginatedResults := filteredApplications[offset:end]

	return paginatedResults, total, nil
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
