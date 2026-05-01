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
	"sync"
	"time"
	"tyforms/internal/database"
	"tyforms/internal/models"
)

const sessionTokenDuration = 30 * 24 * time.Hour

// ApplicationHandler handles all application-related HTTP requests
type ApplicationHandler struct {
	store         *database.SQLiteStore
	adminPassword string
	sessions      map[string]time.Time
	mu            sync.Mutex
}

// NewApplicationHandler creates a new ApplicationHandler
func NewApplicationHandler(store *database.SQLiteStore, adminPassword string) *ApplicationHandler {
	h := &ApplicationHandler{
		store:         store,
		adminPassword: adminPassword,
		sessions:      make(map[string]time.Time),
	}
	go h.cleanupSessions()
	return h
}

func (h *ApplicationHandler) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *ApplicationHandler) createSession() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	token := h.generateToken()
	h.sessions[token] = time.Now().Add(sessionTokenDuration)
	return token
}

func (h *ApplicationHandler) isValidToken(token string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	expiry, ok := h.sessions[token]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		delete(h.sessions, token)
		return false
	}
	return true
}

func (h *ApplicationHandler) revokeToken(token string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, token)
}

func (h *ApplicationHandler) cleanupSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		h.mu.Lock()
		now := time.Now()
		for token, expiry := range h.sessions {
			if now.After(expiry) {
				delete(h.sessions, token)
			}
		}
		h.mu.Unlock()
	}
}

func (h *ApplicationHandler) checkAuth(password, token string) bool {
	if password != "" && password == h.adminPassword {
		return true
	}
	if token != "" && h.isValidToken(token) {
		return true
	}
	return false
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
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

	if !h.checkAuth(request.Password, request.Token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, err := h.store.GetApplication(request.ID)
	if err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	app.IsReviewed = true
	app.ReviewedAt = &time.Time{}
	*app.ReviewedAt = time.Now().UTC()
	app.ReviewNotes = &request.Notes
	app.AcceptanceStatus = request.AcceptanceStatus

	if err := h.store.UpdateApplication(app); err != nil {
		http.Error(w, "Error updating application", http.StatusInternalServerError)
		return
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

	if !h.checkAuth(request.Password, request.Token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, err := h.store.GetApplication(request.ID)
	if err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	app.IsReviewed = false
	app.ReviewedAt = nil
	app.ReviewNotes = nil
	app.AcceptanceStatus = "pending"

	if err := h.store.UpdateApplication(app); err != nil {
		http.Error(w, "Error updating application", http.StatusInternalServerError)
		return
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

	if !h.checkAuth(request.Password, request.Token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.store.DeleteApplication(request.ID); err != nil {
		http.Error(w, "Error deleting application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// VerifyPassword handles password verification requests and returns a session token
func (h *ApplicationHandler) VerifyPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var auth struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	success := auth.Password == h.adminPassword
	response := struct {
		Success bool   `json:"success"`
		Token   string `json:"token,omitempty"`
	}{
		Success: success,
	}

	if success {
		response.Token = h.createSession()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateToken checks if a session token is still valid
func (h *ApplicationHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	valid := h.isValidToken(req.Token)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"valid": valid})
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
