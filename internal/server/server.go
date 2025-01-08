package server

import (
	"GO_FINAL_PROJECT/config"
	"GO_FINAL_PROJECT/internal/database"
	"log"
	"net/http"
)

var JWTSecretKey []byte

// StartServer starts the server
// It serves the static files from the web directory
//
// # It also adds handlers for the API endpoints
//
// The API endpoints are:
// /api/nextdate
// /api/task
// /api/tasks
// /api/task/done
//
// It listens on the port specified in the config
func StartServer(cfg *config.Config, s *database.Service) error {
	// Create file server
	server := http.FileServer(http.Dir(cfg.WebDir))

	// Strip the prefix from the URL
	http.Handle("/", http.StripPrefix("/", server))

	// Add handlers without middleware
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/signin", func(w http.ResponseWriter, r *http.Request) {
		SignInHandler(w, r, cfg.Pass)
	})

	// Add handlers with AuthMiddleware
	http.HandleFunc("/api/task", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		TaskHandler(w, r, s)
	}, cfg.Pass))
	http.HandleFunc("/api/tasks", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		TasksHandler(w, r, s)
	}, cfg.Pass))
	http.HandleFunc("/api/task/done", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		TaskHandlerDone(w, r, s)
	}, cfg.Pass))

	// Start the server
	log.Printf("Starting server on port %s...", cfg.WebPort)
	log.Printf("Open in browser http://localhost:%s", cfg.WebPort)
	if err := http.ListenAndServe(":"+cfg.WebPort, nil); err != nil {
		return err
	}
	return nil
}
