package debughttp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"yanm/internal/config" // Main application config
)

// Config holds the configuration for the debug HTTP server.
type Config struct {
	ListenAddress string
}

// Server represents the debug HTTP server.
type Server struct {
	httpServer         *http.Server
	logger             *slog.Logger
	appConfig          *config.Configuration // Reference to the main application's config
	mux                *http.ServeMux
	registeredHandlers map[string]string // path -> description
}

// NewServer creates and configures a new debug HTTP server.
// It takes the debug server's configuration, the main application's configuration, and a logger.
func NewServer(cfg Config, mainAppConfig *config.Configuration, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	s := &Server{
		logger:             logger,
		appConfig:          mainAppConfig,
		mux:                mux,
		registeredHandlers: make(map[string]string),
	}

	// Setup default handlers
	s.RegisterHandler("/debug/config", "Shows the current application configuration (JSON)", http.HandlerFunc(s.handleDebugConfig))
	s.RegisterHandler("/", "Shows this help page with available debug endpoints", http.HandlerFunc(s.handleRoot))

	s.httpServer = &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: mux,
	}
	return s
}

// RegisterHandler allows other packages to add their own handlers to the debug server.
func (s *Server) RegisterHandler(path string, description string, handler http.Handler) {
	if handler == nil {
		return
	}

	if path == "" || path[0] != '/' {
		s.logger.Warn("Debug handler path must begin with '/'", "path", path)
		// Optionally, prefix with '/' or return an error
		return
	}
	s.mux.Handle(path, handler)
	s.registeredHandlers[path] = description
	s.logger.Info("Registered new debug handler", "path", path, "description", description)
}

// Start runs the debug HTTP server in a new goroutine.
func (s *Server) Start() {
	s.logger.Info("Starting debug HTTP server", "address", s.httpServer.Addr)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Debug HTTP server failed or unexpectedly shut down", "error", err)
		}
	}()
}

// Stop gracefully shuts down the debug HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping debug HTTP server...")
	// Create a context with a timeout for the shutdown, in case the provided context doesn't have one.
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(shutdownCtx)
}

func (s *Server) handleDebugConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.appConfig); err != nil {
		s.logger.Error("Failed to encode config for debug endpoint", "error", err)
		http.Error(w, "Failed to encode configuration", http.StatusInternalServerError)
	}
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Debug HTTP Server is running.\n\nAvailable endpoints:\n")
	// List registered handlers
	for path, description := range s.registeredHandlers {
		fmt.Fprintf(w, "  - %s  (%s)\n", path, description)
	}
}
