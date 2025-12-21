package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"tyforms/internal/config"
	"tyforms/internal/database"
	"tyforms/internal/handlers"
)

// spaHandler serves the SPA and handles client-side routing
type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the absolute path to prevent directory traversal
	path := filepath.Join(h.staticPath, r.URL.Path)

	// Check if the path is a directory
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		// If it's a directory, try to serve index.html from that directory
		indexFile := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexFile); err == nil {
			path = indexFile
		} else {
			// If no index.html in the directory, serve the main index.html
			path = filepath.Join(h.staticPath, h.indexPath)
		}
	} else if os.IsNotExist(err) {
		// File does not exist, serve index.html for SPA routing
		path = filepath.Join(h.staticPath, h.indexPath)
	}

	http.ServeFile(w, r, path)
}

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

	// Create a new ServeMux for better route handling
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/auth/verify", handler.VerifyPassword)
	mux.HandleFunc("/api/application", handler.CreateApplication)
	mux.HandleFunc("/api/application/list", handler.GetApplications)
	mux.HandleFunc("/api/application/export", handler.ExportApplications)
	mux.HandleFunc("/api/application/review", handler.ReviewApplication)
	mux.HandleFunc("/api/application/delete", handler.DeleteApplication)
	mux.HandleFunc("/api/application/unreview", handler.UnreviewApplication)

	// SPA handler for all other routes
	spa := spaHandler{staticPath: "wwwroot", indexPath: "index.html"}
	mux.Handle("/", spa)

	// Start server
	log.Printf("Server starting on port %d", cfg.Server.Port)
	if err := http.ListenAndServe(":5099", mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
