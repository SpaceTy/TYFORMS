package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"tyforms/internal/database"
	"tyforms/internal/models"
)

// ApplicationHandler handles all application-related HTTP requests
type ApplicationHandler struct {
	store         *database.SQLiteStore
	adminPassword string
}

// NewApplicationHandler creates a new ApplicationHandler
func NewApplicationHandler(store *database.SQLiteStore, adminPassword string) *ApplicationHandler {
	return &ApplicationHandler{
		store:         store,
		adminPassword: adminPassword,
	}
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
		http.Error(w, "Error creating application", http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully created application with ID: %d", app.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

// GetApplications handles retrieving all applications (admin only)
func (h *ApplicationHandler) GetApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check admin password
	var auth struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if auth.Password != h.adminPassword {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	applications, err := h.store.GetAllApplications()
	if err != nil {
		http.Error(w, "Error retrieving applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(applications)
}

// ExportApplications handles exporting applications to CSV (admin only)
func (h *ApplicationHandler) ExportApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check admin password
	var auth struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if auth.Password != h.adminPassword {
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
		ID               int    `json:"id"`
		Notes            string `json:"notes"`
		AcceptanceStatus string `json:"acceptance_status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Password != h.adminPassword {
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
		ID       int    `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Password != h.adminPassword {
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
		ID       int    `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Password != h.adminPassword {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.store.DeleteApplication(request.ID); err != nil {
		http.Error(w, "Error deleting application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// VerifyPassword handles password verification requests
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

	response := struct {
		Success bool `json:"success"`
	}{
		Success: auth.Password == h.adminPassword,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
