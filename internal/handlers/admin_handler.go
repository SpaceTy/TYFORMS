package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"tyforms/internal/database"
	"tyforms/internal/models"

	"golang.org/x/crypto/bcrypt"
)

var adminUsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// sessionResponse is the payload returned by login and refresh
type sessionResponse struct {
	Success      bool          `json:"success"`
	Token        string        `json:"token"`
	RefreshToken string        `json:"refreshToken"`
	ExpiresAt    time.Time     `json:"expiresAt"`
	Admin        *models.Admin `json:"admin,omitempty"`
}

// Login authenticates an admin account (username + password). For backwards
// compatibility, an empty username validates the legacy configured admin
// password and logs in as the seeded "admin" account. Also serves the old
// /api/auth/verify endpoint.
func (h *ApplicationHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var admin *models.Admin
	if req.Username == "" {
		// Legacy password-only login (kept for compatibility with existing clients)
		if req.Password == "" || req.Password != h.adminPassword {
			h.writeLoginFailure(w)
			return
		}
		stored, err := h.store.GetAdminByUsername(database.SeedAdminUsername)
		if err != nil || stored == nil {
			admin = &models.Admin{ID: 0, Username: database.SeedAdminUsername}
		} else {
			admin = &stored.Admin
		}
	} else {
		stored, err := h.store.GetAdminByUsername(req.Username)
		if err != nil || stored == nil || !stored.IsActive {
			h.writeLoginFailure(w)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte(req.Password)) != nil {
			h.writeLoginFailure(w)
			return
		}
		admin = &stored.Admin
	}

	token, refreshToken, expiresAt, err := h.createSessionPair(admin.ID)
	if err != nil {
		log.Printf("Error creating session pair: %v", err)
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessionResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Admin:        admin,
	})
}

func (h *ApplicationHandler) writeLoginFailure(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]bool{"success": false})
}

// RefreshSession rotates a refresh token into a fresh access + refresh pair
func (h *ApplicationHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
		return
	}

	adminID, ok, err := h.store.GetValidSession(req.RefreshToken, "refresh")
	if err != nil || !ok {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	admin, err := h.store.GetAdminByID(adminID)
	if err != nil || admin == nil || !admin.IsActive {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Rotate: invalidate the used refresh token before issuing the new pair
	if err := h.store.DeleteSession(req.RefreshToken); err != nil {
		log.Printf("Error rotating refresh token: %v", err)
	}

	token, refreshToken, expiresAt, err := h.createSessionPair(admin.ID)
	if err != nil {
		log.Printf("Error creating session pair: %v", err)
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessionResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Admin:        admin,
	})
}

// Logout revokes the presented session tokens
func (h *ApplicationHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token != "" {
		if err := h.store.DeleteSession(req.Token); err != nil {
			log.Printf("Error revoking access token: %v", err)
		}
	}
	if req.RefreshToken != "" {
		if err := h.store.DeleteSession(req.RefreshToken); err != nil {
			log.Printf("Error revoking refresh token: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// ValidateToken checks if an access session token is still valid and returns
// the authenticated admin when it is
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

	response := map[string]interface{}{"valid": false}
	if admin, ok := h.resolveActor("", req.Token); ok {
		response["valid"] = true
		response["admin"] = admin
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListAdmins returns all admin accounts (admin only)
func (h *ApplicationHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
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

	admins, err := h.store.ListAdmins()
	if err != nil {
		http.Error(w, "Error retrieving admins", http.StatusInternalServerError)
		return
	}
	if admins == nil {
		admins = []*models.Admin{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"admins":  admins,
	})
}

// CreateAdmin adds a new admin account (admin only)
func (h *ApplicationHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
		Token    string `json:"token"`
		Username string `json:"username"`
		NewPass  string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.checkAuth(req.Password, req.Token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username := strings.TrimSpace(req.Username)
	if !adminUsernamePattern.MatchString(username) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Username must be 3-32 characters (letters, numbers, _ or -).",
		})
		return
	}
	if len(req.NewPass) < 6 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Password must be at least 6 characters.",
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPass), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	admin, err := h.store.CreateAdmin(username, string(hash))
	if err == database.ErrAdminExists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "An admin with that username already exists.",
		})
		return
	}
	if err != nil {
		http.Error(w, "Error creating admin", http.StatusInternalServerError)
		return
	}

	log.Printf("Admin account %q created", username)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"admin":   admin,
	})
}

// DeleteAdmin removes an admin account (admin only)
func (h *ApplicationHandler) DeleteAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
		Token    string `json:"token"`
		ID       int    `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	actor, ok := h.resolveActor(req.Password, req.Token)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if actor.ID != 0 && actor.ID == req.ID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "You cannot delete your own account."})
		return
	}

	count, err := h.store.CountAdmins()
	if err != nil {
		http.Error(w, "Error counting admins", http.StatusInternalServerError)
		return
	}
	if count <= 1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot delete the last admin account."})
		return
	}

	target, err := h.store.GetAdminByID(req.ID)
	if err != nil || target == nil {
		http.Error(w, "Admin not found", http.StatusNotFound)
		return
	}

	if err := h.store.DeleteAdmin(req.ID); err != nil {
		http.Error(w, "Error deleting admin", http.StatusInternalServerError)
		return
	}

	log.Printf("Admin account %q deleted by %q", target.Username, actor.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// ChangeAdminPassword updates an admin account's password (admin only).
// When no ID is provided the password of the calling account is changed.
func (h *ApplicationHandler) ChangeAdminPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
		Token    string `json:"token"`
		ID       int    `json:"id"`
		NewPass  string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	actor, ok := h.resolveActor(req.Password, req.Token)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if len(req.NewPass) < 6 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password must be at least 6 characters."})
		return
	}

	targetID := req.ID
	targetUsername := ""
	if targetID == 0 {
		if actor.ID == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Specify the id of the account to update."})
			return
		}
		targetID = actor.ID
		targetUsername = actor.Username
	} else {
		target, err := h.store.GetAdminByID(targetID)
		if err != nil || target == nil {
			http.Error(w, "Admin not found", http.StatusNotFound)
			return
		}
		targetUsername = target.Username
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPass), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	if err := h.store.SetAdminPassword(targetID, string(hash)); err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	log.Printf("Password for admin %q updated by %q", targetUsername, actor.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
