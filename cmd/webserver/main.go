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

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the filesystem. If a file is found, it will be served. If not, the
// file located at the index path will be served to handle SPA routing.
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the absolute path to prevent directory traversal
	path := filepath.Join(h.staticPath, r.URL.Path)

	// Check if path exists and whether it's a file or directory
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		// File does not exist, serve index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// Some other error occurred
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If it's a directory, try to serve index.html from that directory
	if fi.IsDir() {
		indexFile := filepath.Join(path, h.indexPath)
		if _, err := os.Stat(indexFile); err != nil {
			// No index.html in directory, serve main index.html
			http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
			return
		}
	}

	// Serve the file
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
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
	http.HandleFunc("/api/auth/verify", handler.VerifyPassword)
	http.HandleFunc("/api/application", handler.CreateApplication)
	http.HandleFunc("/api/application/list", handler.GetApplications)
	http.HandleFunc("/api/application/export", handler.ExportApplications)
	http.HandleFunc("/api/application/review", handler.ReviewApplication)
	http.HandleFunc("/api/application/delete", handler.DeleteApplication)
	http.HandleFunc("/api/application/unreview", handler.UnreviewApplication)

	// Serve static files and handle SPA routing
	spa := spaHandler{staticPath: "wwwroot", indexPath: "index.html"}
	http.Handle("/", spa)

	// Start server
	log.Printf("Server starting on port %d", cfg.Server.Port)
	if err := http.ListenAndServe(":5099", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
