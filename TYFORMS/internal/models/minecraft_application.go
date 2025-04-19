package models

import "time"

// MinecraftApplication represents a player's application to join the SMP
type MinecraftApplication struct {
	ID                     int        `json:"id"`
	DiscordUsername        string     `json:"discord_username"`
	MinecraftUsername      string     `json:"minecraft_username"`
	Age                    int        `json:"age"`
	FavoriteAboutMinecraft string     `json:"favorite_about_minecraft"`
	UnderstandingOfSMP     string     `json:"understanding_of_smp"`
	JoinedDiscord          bool       `json:"joined_discord"`
	SubmissionDate         time.Time  `json:"submission_date"`
	IsReviewed             bool       `json:"is_reviewed"`
	ReviewedAt             *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes            *string    `json:"review_notes,omitempty"`
	AcceptanceStatus       string     `json:"acceptance_status"`
}
