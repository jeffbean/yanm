package debughttp

import (
	"context"
	_ "embed"
	"html/template"
	"log/slog"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Config holds the configuration for the debug HTTP server.
type Config struct {
	ListenAddress string
}

// Server represents the debug HTTP server.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	mux        *http.ServeMux

	mu                 sync.RWMutex
	registeredHandlers map[string]string // path -> description
}

// NewServer creates and configures a new debug HTTP server.
// It takes the debug server's configuration and a logger.
func NewServer(cfg Config, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	s := &Server{
		logger:             logger,
		mux:                mux,
		registeredHandlers: make(map[string]string),
	}

	// Setup default root handler
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

//go:embed debug_root.html
var _debugRootHTMLTemplate string

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("debugRoot").Parse(_debugRootHTMLTemplate)
	if err != nil {
		s.logger.Error("Failed to parse debug root HTML template", "error", err)
		http.Error(w, "Internal server error: could not load debug template", http.StatusInternalServerError)
		return
	}

	type handlerInfo struct {
		Path        string
		Description string
	}

	var handlers []handlerInfo
	s.mu.RLock() // Use RLock as we are only reading
	for path, description := range s.registeredHandlers {
		handlers = append(handlers, handlerInfo{Path: path, Description: description})
	}
	s.mu.RUnlock()

	// Sort handlers by path for consistent ordering
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].Path < handlers[j].Path
	})

	data := struct {
		Handlers []handlerInfo
	}{
		Handlers: handlers,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		s.logger.ErrorContext(r.Context(), "Failed to execute debug root HTML template", "error", err)
		// If headers haven't been sent, try to send an error
		if w.Header().Get("Content-Type") == "" {
			http.Error(w, "Internal server error: could not render debug page", http.StatusInternalServerError)
		}
	}
}
