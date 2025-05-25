package debughttp

import (
	"bytes" // Added for buffer
	"context"
	"embed"
	_ "embed"
	"errors" // For ErrPathAlreadyRegistered
	"fmt"
	htmltemplate "html/template" // Aliased for clarity
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"sync" // Aliased for clarity
	"yanm/internal/debughttp/debughandler"
)

// NavVisibility determines if a debug route should be visible in navigation links.
type NavVisibility int

const (
	// NavDefault defers to default behavior (typically include).
	NavDefault NavVisibility = iota
	// NavExclude explicitly excludes the route from navigation.
	NavExclude
)

// DebugRoute defines a route for a debug page.
type DebugRoute struct {
	Name        string        // optional
	Path        string        // required
	Description string        // optional
	Handler     http.Handler  // required
	Visibility  NavVisibility // Controls inclusion in navigation links
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

// ErrPathAlreadyRegistered is returned when attempting to register a debug page path that is already in use.
var ErrPathAlreadyRegistered = errors.New("debug page path already registered")

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

	m.mu.Lock() // Lock for checking routes and appending
	defer m.mu.Unlock()

	// Check for duplicate path before registering with underlying mux or adding to routes
	for _, existingRoute := range m.routes {
		if existingRoute.Path == route.Path {
			return fmt.Errorf("%w: %s", ErrPathAlreadyRegistered, route.Path)
		}
	}

	// If we proceed here, it means the standard library Handle should not panic for duplicates,
	// as we've already checked. However, for safety and clarity, if Handle *could* panic
	// for other reasons, more robust panic recovery might be needed here in a real library.
	m.mux.Handle(route.Path, route.Handler) // This is net/http.ServeMux.Handle
	m.routes = append(m.routes, route)
	return nil
}

func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.logger.DebugContext(r.Context(),
		"Debug HTTP request",
		"method", r.Method,
		"url", r.URL)

	m.mu.RLock()
	// Determine NavLinks and Title for the current request
	navLinksResult := make([]debughandler.NavLink, 0, len(m.routes))
	currentTitle := "Debug"
	foundTitle := false

	for _, rt := range m.routes {
		// Add to NavLinks only if visibility is not Exclude
		// Default (zero value) or explicit Include will be added.
		if rt.Visibility != NavExclude {
			navLinksResult = append(navLinksResult, debughandler.NavLink{Path: rt.Path, Name: rt.Name})
		}

		// Check if this route matches the current request path to set the title
		// Exact match or prefix match for directory-like paths (ending in /)
		if r.URL.Path == rt.Path || (strings.HasSuffix(rt.Path, "/") && strings.HasPrefix(r.URL.Path, rt.Path)) {
			if rt.Path == "/" {
				currentTitle = "Debug Home"
				foundTitle = true
			} else if !foundTitle || len(rt.Path) > len(currentTitle) { // Prefer more specific match for title
				currentTitle = rt.Name
				foundTitle = true
			}
		}
	}
	m.mu.RUnlock()

	if !foundTitle && r.URL.Path == "/" { // Ensure root always gets its title if not specifically matched (e.g. if no routes yet)
		currentTitle = "Debug Home"
	}

	pageCtxData := debughandler.PageContextData{
		Title:    currentTitle,
		NavLinks: navLinksResult,
	}
	ctxWithData := debughandler.NewContextWithPageData(r.Context(), pageCtxData)
	rWithData := r.WithContext(ctxWithData)

	m.mux.ServeHTTP(w, rWithData)
}

// NewServer creates and configures a new debug HTTP server.
// It takes the debug server's configuration and a logger.
func NewServer(cfg Config, logger *slog.Logger) (*Server, error) {
	mux := &mux{
		mux:    http.NewServeMux(),
		logger: logger, // Use the validated logger
	}
	serverLogger := logger.With("component", "debug_server")

	server := Server{
		mux: mux,
		httpServer: &http.Server{
			Addr:    cfg.ListenAddress,
			Handler: mux, // Use the custom mux
		},
		logger: serverLogger, // Use the component-specific logger for the server itself
	}

	// Setup default handlers
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		logger.Error("Failed to create sub FS for static assets", "error", err)
		return nil, err
	}

	if err := mux.Handle(DebugRoute{
		Path:       "/debug/static/",
		Handler:    http.StripPrefix("/debug/static/", http.FileServer(http.FS(staticSubFS))),
		Visibility: NavExclude, // Exclude static assets from nav links
	}); err != nil {
		return nil, err
	}

	if err := mux.Handle(DebugRoute{
		Path:    "/",
		Name:    "Home",
		Handler: http.HandlerFunc(server.handleRoot),
	}); err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *Server) RegisterPage(route DebugRoute) error {
	return s.mux.Handle(route)
}

// Start runs the debug HTTP server in a new goroutine.
func (s *Server) Start(_ context.Context) {
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

var _rootTemplate = htmltemplate.Must(htmltemplate.New("root").Parse(_debugRootHTMLTemplate))

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	// handleRoot is responsible for rendering the main index page of the debug server.
	// It lists available debug routes.
	// It needs to use the same layout mechanism as other pages created via NewHTMLProducingHandler.

	// 1. Prepare the data for the _rootTemplate (which is the specific content for this page)
	pageContentData := struct {
		Handlers []DebugRoute
	}{
		Handlers: s.mux.routes, // Get the list of registered routes
	}

	// 2. Execute the _rootTemplate to get its HTML content
	var contentBuf bytes.Buffer
	if err := _rootTemplate.Execute(&contentBuf, pageContentData); err != nil {
		s.logger.ErrorContext(r.Context(), "Failed to execute root debug template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 3. Prepare the data for the main layout template (_layoutTmpl from debughandler)
	//    Retrieve PageContextData which includes Title and NavLinks prepared by mux.ServeHTTP.
	pageCtxData, ok := debughandler.PageDataFromContext(r.Context())
	if !ok {
		s.logger.ErrorContext(r.Context(), "PageContextData not found in context for root handler, this is unexpected.")
		pageCtxData = debughandler.PageContextData{
			Title:    "Debug Home",
			NavLinks: []debughandler.NavLink{{Path: "/", Name: "Home"}},
		}
	}

	layoutPageData := debughandler.Page{
		Title:    pageCtxData.Title,
		NavLinks: pageCtxData.NavLinks,
		ContentBody: htmltemplate.HTML(
			htmltemplate.HTMLEscapeString(contentBuf.String()),
		),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := debughandler.ExecuteLayout(w, layoutPageData); err != nil {
		s.logger.ErrorContext(r.Context(), "Failed to execute debug layout template for root", "error", err)
	}
}
