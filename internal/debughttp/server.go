package debughttp

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"yanm/internal/debughttp/debughandler"

	"embed"
	"io/fs"
)

// DebugRoute defines a route for a debug page.
type DebugRoute struct {
	Name        string       // optional
	Path        string       // required
	Description string       // optional
	Handler     http.Handler // required
}

// Config holds the configuration for the debug HTTP server.
type Config struct {
	ListenAddress string
}

// Server represents the debug HTTP server.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	mux        *mux
}

type mux struct {
	mux    *http.ServeMux
	logger *slog.Logger

	mu     sync.RWMutex
	routes []DebugRoute
}

func (m *mux) Handle(route DebugRoute) error {
	if route.Handler == nil {
		return fmt.Errorf("handler is nil, cannot register page")
	}
	if route.Path == "" || route.Path[0] != '/' {
		return fmt.Errorf("debug page path must begin with '/'")
	}
	return m.handleRoute(route)
}

func (m *mux) handleRoute(route DebugRoute) error {
	if !strings.HasSuffix(route.Path, "/") {
		route.Path = fmt.Sprintf("%s/", route.Path)
	}
	if route.Name == "" {
		route.Name = route.Path
	}
	m.mux.Handle(route.Path, route.Handler)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes = append(m.routes, route)
	return nil
}

func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.logger.DebugContext(r.Context(),
		"Debug HTTP request",
		"method", r.Method,
		"url", r.URL)
	m.mux.ServeHTTP(w, r)
}

// NewServer creates and configures a new debug HTTP server.
// It takes the debug server's configuration and a logger.
func NewServer(cfg Config, logger *slog.Logger) (*Server, error) {
	mux := &mux{
		mux:    http.NewServeMux(),
		logger: logger,
	}
	server := Server{
		httpServer: &http.Server{
			Addr:    cfg.ListenAddress,
			Handler: mux,
		},
		logger: logger,
	}

	// Setup default handlers
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		logger.Error("Failed to create sub FS for static assets", "error", err)
		return nil, err
	}

	mux.handleRoute(DebugRoute{
		Path:    "/debug/static/",
		Handler: http.StripPrefix("/debug/static/", http.FileServer(http.FS(staticSubFS))),
	})

	mux.handleRoute(DebugRoute{
		Path:    "/",
		Handler: http.HandlerFunc(server.handleRoot),
	})

	return &server, nil
}

func (s *Server) RegisterPage(route DebugRoute) error {
	return s.mux.Handle(route)
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
	return s.httpServer.Shutdown(ctx)
}

//go:embed debug_root.html
var _debugRootHTMLTemplate string

//go:embed static
var staticFS embed.FS

var _rootTemplate = template.Must(template.New("root").Parse(_debugRootHTMLTemplate))

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	debughandler.NewHTMLProducingHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := struct {
			Handlers []DebugRoute
		}{
			Handlers: s.mux.routes,
		}
		_rootTemplate.Execute(w, page)
	}))
}
