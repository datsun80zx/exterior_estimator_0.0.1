package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/datsun80zx/exterior_estimator_0.0.1/internal/db"
	"github.com/datsun80zx/exterior_estimator_0.0.1/internal/handlers"
)

func main() {
	// Database lives next to the binary
	exePath, _ := os.Executable()
	dbDir := filepath.Dir(exePath)
	dbPath := filepath.Join(dbDir, "estimator.db")

	// Allow override via env
	if envDB := os.Getenv("ESTIMATOR_DB"); envDB != "" {
		dbPath = envDB
	}

	// For development, use current directory
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbPath = "estimator.db"
	}

	log.Printf("Using database: %s", dbPath)

	// Initialize database
	store, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.DB.Close()

	// Seed default materials if empty
	if err := store.SeedIfEmpty(); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Template directory - check a few common locations
	templateDir := findDir("templates", []string{
		"templates",
		"./templates",
		filepath.Join(filepath.Dir(exePath), "templates"),
	})

	// Initialize handlers
	h, err := handlers.New(store, templateDir)
	if err != nil {
		log.Fatalf("Failed to initialize handlers: %v", err)
	}

	// Routes
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Static files
	staticDir := findDir("static", []string{
		"static",
		"./static",
		filepath.Join(filepath.Dir(exePath), "static"),
	})
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Start server
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	log.Printf("Starting estimator at http://%s", addr)
	log.Printf("Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func findDir(name string, candidates []string) string {
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	log.Fatalf("Could not find %s directory", name)
	return ""
}
