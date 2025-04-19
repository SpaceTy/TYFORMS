package main

import (
	"log"
	"net/http"
	"tyforms/internal/config"
	"tyforms/internal/database"
	"tyforms/internal/handlers"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize database
	store, err := database.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer store.Close()

	// Initialize handler
	handler := handlers.NewApplicationHandler(store, cfg.Admin.Password)

	// Set up routes
	http.HandleFunc("/api/application", handler.CreateApplication)
	http.HandleFunc("/api/application/list", handler.GetApplications)
	http.HandleFunc("/api/application/export", handler.ExportApplications)
	http.HandleFunc("/api/application/review", handler.ReviewApplication)
	http.HandleFunc("/api/application/delete", handler.DeleteApplication)

	// Serve static files from the wwwroot directory
	http.Handle("/", http.FileServer(http.Dir("wwwroot")))

	// Start server
	log.Printf("Server starting on port %d", cfg.Server.Port)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
