package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"tyforms/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// SeedAdminUsername is the default username of the admin account seeded from config
const SeedAdminUsername = "admin"

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
	queries := []string{
		`CREATE TABLE IF NOT EXISTS applications (
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
		)`,
		`CREATE TABLE IF NOT EXISTS admins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			can_manage_admins BOOLEAN NOT NULL DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS admin_sessions (
			token TEXT PRIMARY KEY,
			admin_id INTEGER NOT NULL,
			kind TEXT NOT NULL DEFAULT 'access',
			created_at DATETIME NOT NULL,
			expires_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_sessions_admin ON admin_sessions(admin_id)`,
		`CREATE TABLE IF NOT EXISTS change_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			application_id INTEGER NOT NULL,
			parent_id INTEGER,
			root_id INTEGER,
			admin_id INTEGER,
			admin_username TEXT NOT NULL,
			action TEXT NOT NULL,
			changes TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_change_log_application ON change_log(application_id)`,
		`CREATE INDEX IF NOT EXISTS idx_change_log_root ON change_log(root_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error executing table creation: %w", err)
		}
	}

	// Add can_manage_admins to databases created before the column existed
	// (a no-op when the column is already present)
	if _, err := db.Exec(`ALTER TABLE admins ADD COLUMN can_manage_admins BOOLEAN NOT NULL DEFAULT FALSE`); err != nil &&
		!strings.Contains(err.Error(), "duplicate column name") {
		return fmt.Errorf("error adding can_manage_admins column: %w", err)
	}
	return nil
}

// EnsureSeedAdmin seeds the initial root admin account using the configured
// username and admin password when no admin accounts exist yet. This keeps
// existing deployments working after the upgrade.
func (s *SQLiteStore) EnsureSeedAdmin(username, password string) error {
	if username == "" || password == "" {
		return nil
	}

	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM admins`).Scan(&count); err != nil {
		return fmt.Errorf("error counting admins: %w", err)
	}

	if count > 0 {
		// The root admin always retains admin-management permission
		if _, err := s.db.Exec(`UPDATE admins SET can_manage_admins = TRUE WHERE username = ?`, username); err != nil {
			return fmt.Errorf("error ensuring root admin permissions: %w", err)
		}
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing seed admin password: %w", err)
	}

	_, err = s.db.Exec(
		`INSERT INTO admins (username, password_hash, created_at, is_active, can_manage_admins) VALUES (?, ?, ?, TRUE, TRUE)`,
		username, string(hash), time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("error creating seed admin: %w", err)
	}

	log.Printf("Seeded initial root admin account %q", username)
	return nil
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

// GetApplicationStatistics retrieves aggregate statistics for admin reporting.
func (s *SQLiteStore) GetApplicationStatistics() (*models.ApplicationStatistics, error) {
	stats := &models.ApplicationStatistics{}

	aggregateQuery := `
	SELECT
		COUNT(*) AS total_applications,
		COALESCE(SUM(CASE WHEN is_reviewed = 1 THEN 1 ELSE 0 END), 0) AS reviewed_applications,
		COALESCE(SUM(CASE WHEN is_reviewed = 0 THEN 1 ELSE 0 END), 0) AS unreviewed_applications,
		COALESCE(SUM(CASE WHEN acceptance_status = 'accepted' THEN 1 ELSE 0 END), 0) AS accepted_applications,
		COALESCE(SUM(CASE WHEN acceptance_status = 'pending' THEN 1 ELSE 0 END), 0) AS pending_applications,
		COALESCE(SUM(CASE WHEN acceptance_status = 'rejected' THEN 1 ELSE 0 END), 0) AS rejected_applications,
		COALESCE(SUM(CASE WHEN joined_discord = 1 THEN 1 ELSE 0 END), 0) AS joined_discord_count,
		COALESCE(AVG(age), 0) AS average_age
	FROM applications`

	if err := s.db.QueryRow(aggregateQuery).Scan(
		&stats.TotalApplications,
		&stats.ReviewedApplications,
		&stats.UnreviewedApplications,
		&stats.AcceptedApplications,
		&stats.PendingApplications,
		&stats.RejectedApplications,
		&stats.JoinedDiscordCount,
		&stats.AverageAge,
	); err != nil {
		return nil, fmt.Errorf("error getting aggregate statistics: %w", err)
	}

	type dailyRow struct {
		Date  string
		Count int
	}

	recentRows := make(map[string]int)
	recentQuery := `
	SELECT DATE(submission_date) AS submission_day, COUNT(*) AS total
	FROM applications
	WHERE submission_date >= datetime('now', '-6 days')
	GROUP BY submission_day
	ORDER BY submission_day ASC`

	rows, err := s.db.Query(recentQuery)
	if err != nil {
		return nil, fmt.Errorf("error getting recent submissions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var row dailyRow
		if err := rows.Scan(&row.Date, &row.Count); err != nil {
			return nil, fmt.Errorf("error scanning recent submission row: %w", err)
		}
		recentRows[row.Date] = row.Count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recent submission rows: %w", err)
	}

	stats.RecentSubmissions = make([]models.DailySubmissionCount, 0, 7)
	now := time.Now().UTC()
	for dayOffset := 6; dayOffset >= 0; dayOffset-- {
		currentDay := now.AddDate(0, 0, -dayOffset).Format("2006-01-02")
		stats.RecentSubmissions = append(stats.RecentSubmissions, models.DailySubmissionCount{
			Date:  currentDay,
			Count: recentRows[currentDay],
		})
	}

	return stats, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
