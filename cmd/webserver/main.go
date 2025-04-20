package main

import (
	"log"
	"net/http"
	"tyforms/internal/config"
	"tyforms/internal/database"
	"tyforms/internal/handlers"
)

func main() {
	log.Printf("Starting server initialization")

	// Load configuration
	cfg, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	log.Printf("Configuration loaded successfully")

	// Initialize database
	log.Printf("Initializing database connection to: %s", cfg.Database.Path)
	store, err := database.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer store.Close()
	log.Printf("Database connection established successfully")

	// Initialize handler
	log.Printf("Initializing application handler")
	handler := handlers.NewApplicationHandler(store, cfg.Admin.Password)

	// Set up routes
	log.Printf("Setting up HTTP routes")
	http.HandleFunc("/api/auth/verify", handler.VerifyPassword)
	http.HandleFunc("/api/application", handler.CreateApplication)
	http.HandleFunc("/api/application/list", handler.GetApplications)
	http.HandleFunc("/api/application/export", handler.ExportApplications)
	http.HandleFunc("/api/application/review", handler.ReviewApplication)
	http.HandleFunc("/api/application/delete", handler.DeleteApplication)
	http.HandleFunc("/api/application/unreview", handler.UnreviewApplication)

	// Serve static files from the wwwroot directory
	http.Handle("/", http.FileServer(http.Dir("wwwroot")))

	// Start server
	log.Printf("Server starting on port %d", cfg.Server.Port)
	if err := http.ListenAndServe(":5099", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
