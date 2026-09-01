package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tyforms/internal/database"
	"tyforms/internal/models"
)

const (
	accessTokenDuration  = 1 * time.Hour
	refreshTokenDuration = 30 * 24 * time.Hour
)

// ApplicationHandler handles all application-related HTTP requests
type ApplicationHandler struct {
	store         *database.SQLiteStore
	rootUsername  string
	adminPassword string
}

// NewApplicationHandler creates a new ApplicationHandler
func NewApplicationHandler(store *database.SQLiteStore, rootUsername, adminPassword string) *ApplicationHandler {
	if rootUsername == "" {
		rootUsername = database.SeedAdminUsername
	}
	h := &ApplicationHandler{
		store:         store,
		rootUsername:  rootUsername,
		adminPassword: adminPassword,
	}
	go h.cleanupSessions()
	return h
}

func (h *ApplicationHandler) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *ApplicationHandler) cleanupSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		if err := h.store.DeleteExpiredSessions(); err != nil {
			log.Printf("Error cleaning up expired sessions: %v", err)
		}
	}
}

// resolveActor authenticates a request and returns the admin responsible for it.
// Session tokens map to their admin account; the legacy config password maps to
// the seeded root admin. Returns nil when authentication fails.
func (h *ApplicationHandler) resolveActor(password, token string) (*models.Admin, bool) {
	if token != "" {
		if adminID, ok, err := h.store.GetValidSession(token, "access"); err == nil && ok {
			admin, err := h.store.GetAdminByID(adminID)
			if err == nil && admin != nil && admin.IsActive {
				return admin, true
			}
		}
	}
	if password != "" && password == h.adminPassword {
		if admin, err := h.store.GetAdminByUsername(h.rootUsername); err == nil && admin != nil {
			return &admin.Admin, true
		}
		return &models.Admin{ID: 0, Username: h.rootUsername}, true
	}
	return nil, false
}

func (h *ApplicationHandler) checkAuth(password, token string) bool {
	_, ok := h.resolveActor(password, token)
	return ok
}

// createSessionPair issues a new access + refresh token pair for an admin
func (h *ApplicationHandler) createSessionPair(adminID int) (access, refresh string, expiresAt time.Time, err error) {
	access = h.generateToken()
	refresh = h.generateToken()
	expiresAt = time.Now().UTC().Add(accessTokenDuration)

	if err = h.store.CreateSession(access, adminID, "access", expiresAt); err != nil {
		return "", "", time.Time{}, err
	}
	if err = h.store.CreateSession(refresh, adminID, "refresh", time.Now().UTC().Add(refreshTokenDuration)); err != nil {
		return "", "", time.Time{}, err
	}
	return access, refresh, expiresAt, nil
}

// CreateApplication handles the creation of a new application
func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received application submission request")

	// Read and log the raw request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}
	log.Printf("Raw request body: %s", string(body))

	// Create a new reader for the JSON decoder
	r.Body = io.NopCloser(bytes.NewReader(body))

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var app models.MinecraftApplication
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log the decoded application data
	appJSON, _ := json.MarshalIndent(app, "", "  ")
	log.Printf("Decoded application data:\n%s", string(appJSON))
	log.Printf("Successfully decoded application for user: %s", app.MinecraftUsername)

	// Set submission date to current time
	app.SubmissionDate = time.Now().UTC()
	app.IsReviewed = false
	app.AcceptanceStatus = "pending"

	log.Printf("Attempting to create application in database")
	if err := h.store.CreateApplication(&app); err != nil {
		log.Printf("Database error creating application: %v", err)
		// Check if it's a unique constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed: applications.minecraft_username") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "A player with this Minecraft username has already applied. Please contact an admin if you believe this is an error.",
			})
			return
		}
		http.Error(w, "Error creating application", http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully created application with ID: %d", app.ID)

	// Root node of this application's change tree
	appSnapshot := marshalApp(&app)
	h.recordChange(app.ID, nil, models.ChangeActionCreate, []models.FieldChange{
		{Field: "application", New: &appSnapshot},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

// recordChange persists a change tree node, logging (but not failing the
// request) when recording itself fails
func (h *ApplicationHandler) recordChange(applicationID int, admin *models.Admin, action string, changes []models.FieldChange) {
	var adminID *int
	username := "system"
	if admin != nil {
		id := admin.ID
		adminID = &id
		username = admin.Username
	}
	if err := h.store.RecordChange(applicationID, adminID, username, action, changes); err != nil {
		log.Printf("Error recording %s change for application %d: %v", action, applicationID, err)
	}
}

// marshalApp serializes an application for storage in a change entry
func marshalApp(app *models.MinecraftApplication) string {
	data, err := json.Marshal(app)
	if err != nil {
		return fmt.Sprintf(`{"id": %d}`, app.ID)
	}
	return string(data)
}

// stringPtr returns a pointer to the string value
func stringPtr(s string) *string { return &s }

// diffApplications compares two application states and returns the changed fields
func diffApplications(before, after *models.MinecraftApplication) []models.FieldChange {
	var changes []models.FieldChange

	check := func(field string, oldVal, newVal string) {
		if oldVal != newVal {
			changes = append(changes, models.FieldChange{
				Field: field,
				Old:   stringPtr(oldVal),
				New:   stringPtr(newVal),
			})
		}
	}
	deref := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}
	formatT := func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return t.Format(time.RFC3339)
	}

	check("discordUsername", before.DiscordUsername, after.DiscordUsername)
	check("minecraftUsername", before.MinecraftUsername, after.MinecraftUsername)
	check("age", strconv.Itoa(before.Age), strconv.Itoa(after.Age))
	check("favoriteAboutMinecraft", before.FavoriteAboutMinecraft, after.FavoriteAboutMinecraft)
	check("understandingOfSMP", before.UnderstandingOfSMP, after.UnderstandingOfSMP)
	check("joinedDiscord", fmt.Sprintf("%v", before.JoinedDiscord), fmt.Sprintf("%v", after.JoinedDiscord))
	check("isReviewed", fmt.Sprintf("%v", before.IsReviewed), fmt.Sprintf("%v", after.IsReviewed))
	check("reviewedAt", formatT(before.ReviewedAt), formatT(after.ReviewedAt))
	check("reviewNotes", deref(before.ReviewNotes), deref(after.ReviewNotes))
	check("acceptanceStatus", before.AcceptanceStatus, after.AcceptanceStatus)

	return changes
}

// GetApplications handles retrieving all applications with search and pagination (admin only)
func (h *ApplicationHandler) GetApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request with pagination and search parameters
	var request struct {
		Password string   `json:"password"`
		Token    string   `json:"token"`
		Query    string   `json:"query"`
		Fields   []string `json:"fields"`
		Page     int      `json:"page"`
		PageSize int      `json:"pageSize"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check admin password or token
	if !h.checkAuth(request.Password, request.Token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set defaults
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = 50
	}

	// Validate and whitelist search fields
	allowedFields := map[string]bool{
		"discordUsername":        true,
		"minecraftUsername":      true,
		"favoriteAboutMinecraft": true,
		"understandingOfSMP":     true,
		"id":                     true,
	}

	// Filter out invalid fields
	validFields := []string{}
	for _, field := range request.Fields {
		if allowedFields[field] {
			validFields = append(validFields, field)
		}
	}

	// Call SearchApplications with pagination and search
	applications, total, err := h.store.SearchApplications(request.Query, validFields, request.Page, request.PageSize)
	if err != nil {
		http.Error(w, "Error retrieving applications", http.StatusInternalServerError)
		return
	}

	// Calculate total pages
	totalPages := (total + request.PageSize - 1) / request.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// Build response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":    true,
		"data":       applications,
		"total":      total,
		"page":       request.Page,
		"pageSize":   request.PageSize,
		"totalPages": totalPages,
	}
	json.NewEncoder(w).Encode(response)
}

// GetApplicationStatistics handles retrieving aggregate application statistics (admin only).
func (h *ApplicationHandler) GetApplicationStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var auth struct {
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.checkAuth(auth.Password, auth.Token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.store.GetApplicationStatistics()
	if err != nil {
		http.Error(w, "Error retrieving application statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"data":    stats,
	}
	json.NewEncoder(w).Encode(response)
}

// ExportApplications handles exporting applications to CSV (admin only)
func (h *ApplicationHandler) ExportApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check admin password or token
	var auth struct {
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.checkAuth(auth.Password, auth.Token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	applications, err := h.store.GetAllApplications()
	if err != nil {
		http.Error(w, "Error retrieving applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=applications.csv")

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"ID", "Discord Username", "Minecraft Username", "Age",
		"Favorite About Minecraft", "Understanding of SMP", "Joined Discord",
		"Submission Date", "Is Reviewed", "Reviewed At", "Review Notes",
		"Acceptance Status",
	}
	if err := csvWriter.Write(header); err != nil {
		http.Error(w, "Error writing CSV", http.StatusInternalServerError)
		return
	}

	// Write data
	for _, app := range applications {
		row := []string{
			strconv.Itoa(app.ID),
			app.DiscordUsername,
			app.MinecraftUsername,
			strconv.Itoa(app.Age),
			app.FavoriteAboutMinecraft,
			app.UnderstandingOfSMP,
			fmt.Sprintf("%v", app.JoinedDiscord),
			app.SubmissionDate.Format(time.RFC3339),
			fmt.Sprintf("%v", app.IsReviewed),
			formatTime(app.ReviewedAt),
			formatString(app.ReviewNotes),
			app.AcceptanceStatus,
		}
		if err := csvWriter.Write(row); err != nil {
			http.Error(w, "Error writing CSV", http.StatusInternalServerError)
			return
		}
	}
}

// ReviewApplication handles reviewing an application (admin only)
func (h *ApplicationHandler) ReviewApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Password         string `json:"password"`
		Token            string `json:"token"`
		ID               int    `json:"id"`
		Notes            string `json:"notes"`
		AcceptanceStatus string `json:"acceptance_status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	actor, ok := h.resolveActor(request.Password, request.Token)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, err := h.store.GetApplication(request.ID)
	if err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}
	before := *app

	app.IsReviewed = true
	app.ReviewedAt = &time.Time{}
	*app.ReviewedAt = time.Now().UTC()
	app.ReviewNotes = &request.Notes
	app.AcceptanceStatus = request.AcceptanceStatus

	if err := h.store.UpdateApplication(app); err != nil {
		http.Error(w, "Error updating application", http.StatusInternalServerError)
		return
	}

	if diff := diffApplications(&before, app); len(diff) > 0 {
		h.recordChange(app.ID, actor, models.ChangeActionReview, diff)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// UnreviewApplication handles unreviewing an application (admin only)
func (h *ApplicationHandler) UnreviewApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Password string `json:"password"`
		Token    string `json:"token"`
		ID       int    `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	actor, ok := h.resolveActor(request.Password, request.Token)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, err := h.store.GetApplication(request.ID)
	if err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}
	before := *app

	app.IsReviewed = false
	app.ReviewedAt = nil
	app.ReviewNotes = nil
	app.AcceptanceStatus = "pending"

	if err := h.store.UpdateApplication(app); err != nil {
		http.Error(w, "Error updating application", http.StatusInternalServerError)
		return
	}

	if diff := diffApplications(&before, app); len(diff) > 0 {
		h.recordChange(app.ID, actor, models.ChangeActionUnreview, diff)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// DeleteApplication handles deleting an application (admin only)
func (h *ApplicationHandler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Password string `json:"password"`
		Token    string `json:"token"`
		ID       int    `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	actor, ok := h.resolveActor(request.Password, request.Token)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, err := h.store.GetApplication(request.ID)
	if err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	if err := h.store.DeleteApplication(request.ID); err != nil {
		http.Error(w, "Error deleting application", http.StatusInternalServerError)
		return
	}

	h.recordChange(request.ID, actor, models.ChangeActionDelete, []models.FieldChange{
		{Field: "application", Old: stringPtr(marshalApp(app))},
	})

	w.WriteHeader(http.StatusOK)
}

// UpdateApplication handles updating application content (admin only)
func (h *ApplicationHandler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Password               string  `json:"password"`
		Token                  string  `json:"token"`
		ID                     int     `json:"id"`
		DiscordUsername        *string `json:"discordUsername"`
		MinecraftUsername      *string `json:"minecraftUsername"`
		Age                    *int    `json:"age"`
		FavoriteAboutMinecraft *string `json:"favoriteAboutMinecraft"`
		UnderstandingOfSMP     *string `json:"understandingOfSMP"`
		JoinedDiscord          *bool   `json:"joinedDiscord"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	actor, ok := h.resolveActor(request.Password, request.Token)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, err := h.store.GetApplication(request.ID)
	if err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}
	before := *app

	if request.DiscordUsername != nil {
		app.DiscordUsername = *request.DiscordUsername
	}
	if request.MinecraftUsername != nil {
		app.MinecraftUsername = *request.MinecraftUsername
	}
	if request.Age != nil {
		app.Age = *request.Age
	}
	if request.FavoriteAboutMinecraft != nil {
		app.FavoriteAboutMinecraft = *request.FavoriteAboutMinecraft
	}
	if request.UnderstandingOfSMP != nil {
		app.UnderstandingOfSMP = *request.UnderstandingOfSMP
	}
	if request.JoinedDiscord != nil {
		app.JoinedDiscord = *request.JoinedDiscord
	}

	if err := h.store.UpdateApplication(app); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: applications.minecraft_username") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "A player with this Minecraft username already exists.",
			})
			return
		}
		http.Error(w, "Error updating application", http.StatusInternalServerError)
		return
	}

	if diff := diffApplications(&before, app); len(diff) > 0 {
		h.recordChange(app.ID, actor, models.ChangeActionUpdate, diff)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// GetApplicationHistory returns the change tree for one application (admin only)
func (h *ApplicationHandler) GetApplicationHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Password string `json:"password"`
		Token    string `json:"token"`
		ID       int    `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.checkAuth(request.Password, request.Token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	changes, err := h.store.GetApplicationChanges(request.ID)
	if err != nil {
		http.Error(w, "Error retrieving change history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"changes": changes,
	})
}

// GetRecentChanges returns the latest changes across all applications (admin only)
func (h *ApplicationHandler) GetRecentChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Password string `json:"password"`
		Token    string `json:"token"`
		Limit    int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.checkAuth(request.Password, request.Token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	changes, err := h.store.GetRecentChanges(request.Limit)
	if err != nil {
		http.Error(w, "Error retrieving changes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"changes": changes,
	})
}

// Helper functions
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func formatString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
