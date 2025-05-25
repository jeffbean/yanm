package debughttp

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"yanm/internal/debughttp/debughandler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	srv, err := NewServer(Config{ListenAddress: ":0"}, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	require.NoError(t, err, "NewServer() returned an unexpected error: %v", err)
	require.NotNil(t, srv, "NewServer() returned a nil server")
	require.NotNil(t, srv.httpServer, "Server's httpServer should not be nil")
	require.NotNil(t, srv.logger)
	require.NotNil(t, srv.mux, "Server's mux should not be nil")
}

// TestServer_RegisterPage tests various scenarios for page registration,
// including input validation, path/name normalization, and successful registration.
func TestServer_RegisterPage(t *testing.T) {
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("test content"))
		require.NoError(t, err, "Write() returned an unexpected error: %v", err)
	})

	testCases := []struct {
		name         string
		route        DebugRoute
		expectError  bool
		errorMsg     string // Substring to check in error message if expectError is true
		checkRoute   bool   // Whether to check if the route was added to mux.routes
		expectedPath string // Expected path after normalization (if checkRoute is true)
		expectedName string // Expected name after defaulting (if checkRoute is true)
	}{
		{
			name: "Valid registration",
			route: DebugRoute{
				Name:    "Test Page",
				Path:    "/testpage",
				Handler: baseHandler,
			},
			expectError:  false,
			checkRoute:   true,
			expectedPath: "/testpage/",
			expectedName: "Test Page",
		},
		{
			name: "Nil handler",
			route: DebugRoute{
				Name:    "Nil Handler Page",
				Path:    "/nilhandler",
				Handler: nil,
			},
			expectError: true,
			errorMsg:    "handler is nil",
			checkRoute:  false,
		},
		{
			name: "Empty path",
			route: DebugRoute{
				Name:    "Empty Path Page",
				Path:    "",
				Handler: baseHandler,
			},
			expectError: true,
			errorMsg:    "debug page path must begin with '/'",
			checkRoute:  false,
		},
		{
			name: "Path not starting with slash",
			route: DebugRoute{
				Name:    "No Slash Path",
				Path:    "noslash",
				Handler: baseHandler,
			},
			expectError: true,
			errorMsg:    "debug page path must begin with '/'",
			checkRoute:  false,
		},
		{
			name: "Path normalization - add trailing slash",
			route: DebugRoute{
				Name:    "Needs Slash",
				Path:    "/needsslash", // No trailing slash
				Handler: baseHandler,
			},
			expectError:  false,
			checkRoute:   true,
			expectedPath: "/needsslash/", // Slash should be added
			expectedName: "Needs Slash",
		},
		{
			name: "Path normalization - already has trailing slash",
			route: DebugRoute{
				Name:    "Has Slash",
				Path:    "/hasslash/", // Already has trailing slash
				Handler: baseHandler,
			},
			expectError:  false,
			checkRoute:   true,
			expectedPath: "/hasslash/",
			expectedName: "Has Slash",
		},
		{
			name: "Name defaulting - empty name",
			route: DebugRoute{
				Name:    "", // Empty name
				Path:    "/emptyname",
				Handler: baseHandler,
			},
			expectError:  false,
			checkRoute:   true,
			expectedPath: "/emptyname/",
			expectedName: "/emptyname/", // Name should default to path (after path normalization)
		},
		// Note: Duplicate registration is handled by a separate test TestServer_RegisterPage_DuplicateError
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{ListenAddress: ":0"}
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})) // Reduce noise
			srv, _ := NewServer(cfg, logger)

			err := srv.RegisterPage(tc.route)

			if tc.expectError {
				assert.Error(t, err, "Expected an error for test: %s", tc.name)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg, "Error message mismatch for test: %s", tc.name)
				}
			} else {
				assert.NoError(t, err, "Did not expect an error for test: %s", tc.name)
			}

			if tc.checkRoute {
				found := false
				srv.mux.mu.RLock()
				for _, r := range srv.mux.routes {
					if r.Path == tc.expectedPath && r.Name == tc.expectedName {
						found = true
						break
					}
				}
				srv.mux.mu.RUnlock()
				assert.True(t, found,
					fmt.Sprintf("Registered route not found or incorrect in mux routes. Expected Path: '%s', Name: '%s' for test: %s", tc.expectedPath, tc.expectedName, tc.name)) // Added test name to assert msg
			}
		})
	}
}

// TestServer_RegisterPage_DuplicateError tests that registering the same path twice returns an error.
func TestServer_RegisterPage_DuplicateError(t *testing.T) {
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("test content"))
		require.NoError(t, err, "Write() returned an unexpected error: %v", err)
	})
	cfg := Config{ListenAddress: ":0"}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	srv, _ := NewServer(cfg, logger)

	route1 := DebugRoute{
		Name:    "First Page",
		Path:    "/duplicatepath",
		Handler: baseHandler,
	}
	err1 := srv.RegisterPage(route1)
	require.NoError(t, err1, "Registration of the first page should succeed")

	route2 := DebugRoute{
		Name:    "Second Page",
		Path:    "/duplicatepath", // Same path as route1
		Handler: baseHandler,
	}

	err2 := srv.RegisterPage(route2)
	assert.Error(t, err2, "Registering the same path twice should return an error")
	assert.True(t, errors.Is(err2, ErrPathAlreadyRegistered), "Error should be ErrPathAlreadyRegistered")
	assert.Contains(t, err2.Error(), "/duplicatepath", "Error message should contain the duplicate path")

	// Also test with and without trailing slash consistency
	srv2, _ := NewServer(cfg, logger)
	route3 := DebugRoute{Name: "Path without slash", Path: "/otherduplicate", Handler: baseHandler}
	err3 := srv2.RegisterPage(route3)
	require.NoError(t, err3)

	route4 := DebugRoute{Name: "Path with slash", Path: "/otherduplicate/", Handler: baseHandler}
	err4 := srv2.RegisterPage(route4)
	assert.Error(t, err4, "Registering paths that normalize to the same value should return an error")
	assert.True(t, errors.Is(err4, ErrPathAlreadyRegistered), "Error for normalized path should be ErrPathAlreadyRegistered")
	assert.Contains(t, err4.Error(), "/otherduplicate/", "Error message for normalized path should contain the path")
}

func TestMux_ServeHTTP(t *testing.T) {
	defaultLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})) // Default to error to keep test output clean

	testCases := []struct {
		name               string
		setupMux           func(m *mux) // Function to register routes or configure mux
		reqPath            string
		expectedStatusCode int
		expectedBody       []string // Substrings to check in the body
		expectedHeaders    map[string]string
		unexpectedBody     []string // Substrings that should NOT be in the body
	}{
		{
			name: "Simple HTML handler (no layout)",
			setupMux: func(m *mux) {
				basicHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "text/html; charset=utf-8") // Handler must set its own type
					w.Write([]byte("<p>Hello from basic handler</p>"))
				})
				err := m.Handle(DebugRoute{Name: "Basic", Path: "/basic", Handler: basicHandler})
				require.NoError(t, err)
			},
			reqPath:            "/basic/",
			expectedStatusCode: http.StatusOK,
			// For a plain http.HandlerFunc, no layout is applied, no title injected, no nav.
			expectedBody:    []string{"<p>Hello from basic handler</p>"},
			expectedHeaders: map[string]string{"Content-Type": "text/html; charset=utf-8"},
			unexpectedBody:  []string{"<html>", "<title>Basic</title>", "<nav>"},
		},
		{
			name: "HTMLProducingHandler with layout",
			setupMux: func(m *mux) {
				contentHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte("<h1>Content Page</h1><p>Some details here.</p>"))
				})
				// NewHTMLProducingHandler only takes the source handler.
				// The title is derived from DebugRoute.Name by mux.ServeHTTP.
				producingHandler := debughandler.NewHTMLProducingHandler(contentHandler)
				err := m.Handle(DebugRoute{Name: "Content Page", Path: "/content", Handler: producingHandler})
				require.NoError(t, err)
			},
			reqPath:            "/content/",
			expectedStatusCode: http.StatusOK,
			expectedBody: []string{
				"<html lang=\"en\">", // Updated to match actual output
				"<body>",
				"<title>Content Page - YANM Debug</title>", // Updated to match actual output
				"<nav>",
				"<a href=\"/content/\">Content Page</a>",
				// "<a href=\"/\">Home</a>", // Home link is not registered in this test's mux setup
				"<h1>Content Page</h1><p>Some details here.</p>",
			},
			expectedHeaders: map[string]string{"Content-Type": "text/html; charset=utf-8"},
		},
		{
			name: "Route not found (404)",
			setupMux: func(_ *mux) {
				// No routes registered for this specific test, or a route that won't match
			},
			reqPath:            "/this-path-does-not-exist",
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       []string{"404 page not found"},                                 // Default http.ServeMux message
			unexpectedBody:     []string{"<html>", "<nav>", "YANM Debug"},                      // Layout should not be applied
			expectedHeaders:    map[string]string{"Content-Type": "text/plain; charset=utf-8"}, // Default Go 404 content type
		},
		{
			name: "Root path with HTMLProducingHandler",
			setupMux: func(m *mux) {
				rootContentHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte("<h2>Welcome to Debug Home</h2>"))
				})
				producingHandler := debughandler.NewHTMLProducingHandler(rootContentHandler)
				err := m.Handle(DebugRoute{Name: "Debug Home", Path: "/", Handler: producingHandler})
				require.NoError(t, err)

				// Add another page to see it in nav along with Home
				otherContentHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte("<h1>Other Page</h1>"))
				})
				otherProducingHandler := debughandler.NewHTMLProducingHandler(otherContentHandler)
				err = m.Handle(DebugRoute{Name: "Other Page", Path: "/other", Handler: otherProducingHandler})
				require.NoError(t, err)
			},
			reqPath:            "/",
			expectedStatusCode: http.StatusOK,
			expectedBody: []string{
				"<html lang=\"en\">",
				"<title>Debug Home - YANM Debug</title>", // Specific title for root
				"<nav>",
				"<a href=\"/\">Debug Home</a>",       // Nav link to Home
				"<a href=\"/other/\">Other Page</a>", // Nav link to Other
				"<h2>Welcome to Debug Home</h2>",     // Actual content from root handler
			},
			expectedHeaders: map[string]string{"Content-Type": "text/html; charset=utf-8"},
		},
		{
			name: "HTMLProducingHandler returns error status",
			setupMux: func(m *mux) {
				errorContentHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError) // Set error status
					w.Write([]byte("<h1>Internal Server Error</h1><p>Something went wrong.</p>"))
				})
				producingHandler := debughandler.NewHTMLProducingHandler(errorContentHandler)
				err := m.Handle(DebugRoute{Name: "Error Page", Path: "/errorpage", Handler: producingHandler})
				require.NoError(t, err)
			},
			reqPath:            "/errorpage/",
			expectedStatusCode: http.StatusInternalServerError, // Expect the error status to be preserved
			expectedBody: []string{
				"<html lang=\"en\">",
				"<title>Error Page - YANM Debug</title>", // Title should still be set
				"<nav>",                                  // Layout should still apply
				"<a href=\"/errorpage/\">Error Page</a>",
			},
			expectedHeaders: map[string]string{"Content-Type": "text/html; charset=utf-8"},
		},
		{
			name: "Layout execution error",
			setupMux: func(m *mux) {
				// A simple handler that should try to render through the layout
				contentHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Write([]byte("<h1>Content that may be partially written</h1>"))
				})
				producingHandler := debughandler.NewHTMLProducingHandler(contentHandler)
				err := m.Handle(DebugRoute{Name: "Layout Error Test", Path: "/layout-error", Handler: producingHandler})
				require.NoError(t, err)
			},
			reqPath:            "/layout-error/",
			expectedStatusCode: http.StatusOK, // Status is written before the body write that fails.
			// We can't reliably check the body here as it might be partially written or not at all.
			// The main check is that the handler doesn't panic and status is set.
		},
		// Additional test cases will be added here
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mux{
				mux:    http.NewServeMux(),
				logger: defaultLogger, // Reverted: No longer need test-specific logger for this case
				// routes will be populated by setupMux
			}

			if tc.setupMux != nil {
				tc.setupMux(m)
			}

			req := httptest.NewRequest("GET", tc.reqPath, nil)
			var rr http.ResponseWriter = httptest.NewRecorder() // Default recorder

			if tc.name == "Layout execution error" {
				rr = &failingResponseWriter{ResponseRecorder: httptest.NewRecorder(), failAfterBytes: 50} // Arbitrary small number of bytes
			}

			m.ServeHTTP(rr, req)

			// Check status code using type assertion to get to underlying recorder if necessary
			var finalStatusCode int
			if frw, ok := rr.(*failingResponseWriter); ok {
				finalStatusCode = frw.Code
			} else if rec, ok := rr.(*httptest.ResponseRecorder); ok {
				finalStatusCode = rec.Code
			}
			assert.Equal(t, tc.expectedStatusCode, finalStatusCode, "Status code mismatch")

			if rec, ok := rr.(*httptest.ResponseRecorder); ok { // Only check body for non-failing recorder cases
				bodyStr := rec.Body.String()
				for _, expected := range tc.expectedBody {
					assert.Contains(t, bodyStr, expected, "Expected body content not found: %s", expected)
				}
				for _, unexpected := range tc.unexpectedBody {
					assert.NotContains(t, bodyStr, unexpected, "Unexpected body content found: %s", unexpected)
				}
			}

			// Check headers using type assertion
			var finalHeaders http.Header
			if frw, ok := rr.(*failingResponseWriter); ok {
				finalHeaders = frw.Header()
			} else if rec, ok := rr.(*httptest.ResponseRecorder); ok {
				finalHeaders = rec.Header()
			}
			for key, val := range tc.expectedHeaders {
				assert.Equal(t, val, finalHeaders.Get(key), fmt.Sprintf("Header %s mismatch", key))
			}
		})
	}
}

// failingResponseWriter is a wrapper around httptest.ResponseRecorder to simulate write errors.
// It can be configured to fail after a certain number of bytes have been written.
type failingResponseWriter struct {
	*httptest.ResponseRecorder
	failAfterBytes int // Fail after this many bytes have been successfully written
	bytesWritten   int
	failed         bool // To ensure we only return the error once per conceptual write operation
}

// WriteHeader is a pass-through to the embedded ResponseRecorder.
func (frw *failingResponseWriter) WriteHeader(statusCode int) {
	frw.ResponseRecorder.WriteHeader(statusCode)
}

// Write simulates a write error after failAfterBytes have been written.
func (frw *failingResponseWriter) Write(p []byte) (int, error) {
	if frw.failed {
		return 0, errors.New("simulated write error: already failed")
	}
	if frw.bytesWritten >= frw.failAfterBytes {
		frw.failed = true
		return 0, errors.New("simulated write error: failAfterBytes threshold reached")
	}

	remainingCapacity := frw.failAfterBytes - frw.bytesWritten
	if len(p) > remainingCapacity {
		n, err := frw.ResponseRecorder.Write(p[:remainingCapacity])
		frw.bytesWritten += n
		if err != nil {
			frw.failed = true
			return n, err
		}
		frw.failed = true // Mark as failed because we couldn't write the whole slice p
		return n, errors.New("simulated write error: partial write then threshold reached")
	}

	n, err := frw.ResponseRecorder.Write(p)
	frw.bytesWritten += n
	if err != nil {
		frw.failed = true
	}
	return n, err
}
