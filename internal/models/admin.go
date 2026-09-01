package models

import "time"

// Admin represents an admin account
type Admin struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	CreatedAt       time.Time `json:"createdAt"`
	IsActive        bool      `json:"isActive"`
	CanManageAdmins bool      `json:"canManageAdmins"`
}

// AdminWithHash is used internally when the password hash is needed
type AdminWithHash struct {
	Admin
	PasswordHash string `json:"-"`
}

// FieldChange describes a single field modification within a change entry
type FieldChange struct {
	Field string  `json:"field"`
	Old   *string `json:"old,omitempty"`
	New   *string `json:"new,omitempty"`
}

// ChangeEntry is one node in an application's change tree. Each entry links to
// the change it superseded via ParentID; RootID points at the creation event.
// Concurrent changes sharing the same parent form branches of the tree.
type ChangeEntry struct {
	ID            int           `json:"id"`
	ApplicationID int           `json:"applicationId"`
	ParentID      *int          `json:"parentId"`
	RootID        *int          `json:"rootId"`
	AdminID       *int          `json:"adminId"`
	AdminUsername string        `json:"adminUsername"`
	Action        string        `json:"action"`
	Changes       []FieldChange `json:"changes"`
	CreatedAt     time.Time     `json:"createdAt"`
}

// Actions recorded in the change tree
const (
	ChangeActionCreate   = "create"
	ChangeActionUpdate   = "update"
	ChangeActionReview   = "review"
	ChangeActionUnreview = "unreview"
	ChangeActionDelete   = "delete"
)
