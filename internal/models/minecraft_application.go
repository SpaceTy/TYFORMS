package models

import "time"

// MinecraftApplication represents a player's application to join the SMP
type MinecraftApplication struct {
	ID                     int        `json:"id"`
	DiscordUsername        string     `json:"discordUsername"`
	MinecraftUsername      string     `json:"minecraftUsername"`
	Age                    int        `json:"age"`
	FavoriteAboutMinecraft string     `json:"favoriteAboutMinecraft"`
	UnderstandingOfSMP     string     `json:"understandingOfSMP"`
	JoinedDiscord          bool       `json:"joinedDiscord"`
	SubmissionDate         time.Time  `json:"submissionDate"`
	IsReviewed             bool       `json:"isReviewed"`
	ReviewedAt             *time.Time `json:"reviewedAt,omitempty"`
	ReviewNotes            *string    `json:"reviewNotes,omitempty"`
	AcceptanceStatus       string     `json:"acceptanceStatus"`
}
