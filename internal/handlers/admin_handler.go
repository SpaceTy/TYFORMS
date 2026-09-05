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

// actorCanManageAdmins reports whether the resolved actor may access the
// admins page and manage admin accounts. The legacy config password (which
// resolves to the root admin with ID 0) always may.
func actorCanManageAdmins(actor *models.Admin) bool {
	return actor != nil && (actor.ID == 0 || actor.CanManageAdmins)
}

// requireAdminManager authenticates the request and additionally enforces the
// admin-management permission. Writes the error response and returns nil when
// the caller is not allowed.
func (h *ApplicationHandler) requireAdminManager(w http.ResponseWriter, password, token string) *models.Admin {
	actor, ok := h.resolveActor(password, token)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}
	if !actorCanManageAdmins(actor) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}
	return actor
}

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
		log.Printf("[LOGIN] Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[LOGIN] Attempt for username=%q (legacy=%v) from %s", req.Username, req.Username == "", r.RemoteAddr)

	var admin *models.Admin
	if req.Username == "" {
		// Legacy password-only login (kept for compatibility with existing clients)
		if req.Password == "" {
			log.Printf("[LOGIN] Legacy login failed: empty password")
			h.writeLoginFailure(w)
			return
		}
		if req.Password != h.adminPassword {
			log.Printf("[LOGIN] Legacy login failed: password mismatch (configured password len=%d, supplied len=%d)", len(h.adminPassword), len(req.Password))
			h.writeLoginFailure(w)
			return
		}
		log.Printf("[LOGIN] Legacy password matched, looking up root admin %q", h.rootUsername)
		stored, err := h.store.GetAdminByUsername(h.rootUsername)
		if err != nil {
			log.Printf("[LOGIN] Legacy login: error looking up root admin %q: %v", h.rootUsername, err)
		}
		if err != nil || stored == nil {
			log.Printf("[LOGIN] Legacy login: root admin %q not found in DB, using synthetic admin", h.rootUsername)
			admin = &models.Admin{ID: 0, Username: h.rootUsername}
		} else {
			log.Printf("[LOGIN] Legacy login: root admin found, id=%d active=%v", stored.ID, stored.IsActive)
			admin = &stored.Admin
		}
	} else {
		stored, err := h.store.GetAdminByUsername(req.Username)
		if err != nil {
			log.Printf("[LOGIN] Error looking up admin %q: %v", req.Username, err)
			h.writeLoginFailure(w)
			return
		}
		if stored == nil {
			log.Printf("[LOGIN] Admin %q not found in DB", req.Username)
			h.writeLoginFailure(w)
			return
		}
		if !stored.IsActive {
			log.Printf("[LOGIN] Admin %q exists but is inactive (id=%d)", req.Username, stored.ID)
			h.writeLoginFailure(w)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte(req.Password)) != nil {
			log.Printf("[LOGIN] Password mismatch for admin %q (id=%d, hash_len=%d, supplied_len=%d)", req.Username, stored.ID, len(stored.PasswordHash), len(req.Password))
			h.writeLoginFailure(w)
			return
		}
		log.Printf("[LOGIN] Password verified for admin %q (id=%d)", req.Username, stored.ID)
		admin = &stored.Admin
	}

	token, refreshToken, expiresAt, err := h.createSessionPair(admin.ID)
	if err != nil {
		log.Printf("[LOGIN] Error creating session pair for admin %q (id=%d): %v", admin.Username, admin.ID, err)
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	log.Printf("[LOGIN] Success for admin %q (id=%d)", admin.Username, admin.ID)
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
		log.Printf("[REFRESH] Missing refresh token from %s", r.RemoteAddr)
		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
		return
	}

	adminID, ok, err := h.store.GetValidSession(req.RefreshToken, "refresh")
	if err != nil {
		log.Printf("[REFRESH] Error validating refresh token: %v", err)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}
	if !ok {
		log.Printf("[REFRESH] Refresh token not found or expired from %s", r.RemoteAddr)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	admin, err := h.store.GetAdminByID(adminID)
	if err != nil {
		log.Printf("[REFRESH] Error looking up admin (id=%d): %v", adminID, err)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}
	if admin == nil || !admin.IsActive {
		log.Printf("[REFRESH] Admin (id=%d) not found or inactive", adminID)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Rotate: invalidate the used refresh token before issuing the new pair
	if err := h.store.DeleteSession(req.RefreshToken); err != nil {
		log.Printf("[REFRESH] Error rotating refresh token: %v", err)
	}

	token, refreshToken, expiresAt, err := h.createSessionPair(admin.ID)
	if err != nil {
		log.Printf("[REFRESH] Error creating session pair for admin %q (id=%d): %v", admin.Username, admin.ID, err)
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	log.Printf("[REFRESH] Success for admin %q (id=%d)", admin.Username, admin.ID)
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

// ListAdmins returns all admin accounts (admin-management permission required)
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

	if h.requireAdminManager(w, auth.Password, auth.Token) == nil {
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

// CreateAdmin adds a new admin account (admin-management permission required)
func (h *ApplicationHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password  string `json:"password"`
		Token     string `json:"token"`
		Username  string `json:"username"`
		NewPass   string `json:"newPassword"`
		CanManage bool   `json:"canManageAdmins"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	actor := h.requireAdminManager(w, req.Password, req.Token)
	if actor == nil {
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

	admin, err := h.store.CreateAdmin(username, string(hash), req.CanManage)
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

// DeleteAdmin removes an admin account (admin-management permission required)
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

	actor := h.requireAdminManager(w, req.Password, req.Token)
	if actor == nil {
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

// ChangeAdminPassword updates an admin account's password.
// Requires admin-management permission, except that every account may change
// its own password. When no ID is provided the password of the calling
// account is changed.
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
		// Changing another account's password requires the permission
		if !actorCanManageAdmins(actor) && targetID != actor.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
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

// SetAdminPermissions updates whether an admin may access and manage the
// admins page (admin-management permission required)
func (h *ApplicationHandler) SetAdminPermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password  string `json:"password"`
		Token     string `json:"token"`
		ID        int    `json:"id"`
		CanManage bool   `json:"canManageAdmins"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	actor := h.requireAdminManager(w, req.Password, req.Token)
	if actor == nil {
		return
	}

	if req.ID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Specify the id of the account to update."})
		return
	}

	if actor.ID != 0 && actor.ID == req.ID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "You cannot change your own admin-management permission."})
		return
	}

	target, err := h.store.GetAdminByID(req.ID)
	if err != nil || target == nil {
		http.Error(w, "Admin not found", http.StatusNotFound)
		return
	}

	// The root admin always retains admin-management permission
	if target.Username == h.rootUsername && !req.CanManage {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "The root admin account always retains admin-management permission."})
		return
	}

	if err := h.store.SetAdminPermissions(req.ID, req.CanManage); err != nil {
		http.Error(w, "Error updating permissions", http.StatusInternalServerError)
		return
	}

	log.Printf("Admin-management permission for %q set to %v by %q", target.Username, req.CanManage, actor.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
