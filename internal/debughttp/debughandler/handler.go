package debughandler

import (
	"bytes"
	_ "embed"
	"html/template"
	"io" // Added for io.Writer in ExecuteLayout
	"net/http"
	"net/http/httptest"
)

var (
	//go:embed layout.html
	_layoutHTMLTemplate string

	_layoutTmpl = template.Must(template.New("layout").Parse(_layoutHTMLTemplate))
)

// NavLink represents a navigation link in the debug UI.
// It should be exported so other packages can construct it if needed.
// For now, it's primarily used by the layout template itself.
// Path is the URL, Name is the display name.
// TODO: Consider if this should be populated from the registered routes automatically by the Server.
// For now, pages provide their own if they want to customize (e.g. for breadcrumbs or specific sub-nav).
// The main layout will try to generate a default set from registered routes if Page.NavLinks is empty.
// Update: The layout.html iterates over {{.NavLinks}}. If this is empty, nothing is shown.
// The `server.go` handleRoot will construct these from its known routes.
// Other pages (via NewHTMLProducingHandler) won't have specific NavLinks passed to layout.html this way,
// layout.html currently does *not* have access to server's routes directly.
// This means that only the root page will show the dynamic nav list in the header for now.
// This might be acceptable as individual pages are usually navigated to *from* the root.
// Or the layout could be made more complex to always receive the full route list.

// NavLink defines a link for the navigation bar.
// To be used by the layout template.
// Path is the URL, Name is the link text.
type NavLink struct {
	Path string
	Name string
}

// Page represents the data passed to the layout template.
// It includes the main content and any other metadata needed by the layout.
type Page struct {
	Title       string // Title of the HTML page
	NavLinks    []NavLink // Navigation links for the header/sidebar
	ContentBody template.HTML // The main HTML content for the page
	// RequestContext can be used to pass around values to the template.
	// It is not used by the layout, but can be used by the content.
}

// ExecuteLayout executes the main layout template with the provided Page data,
// writing the output to w.
func ExecuteLayout(w io.Writer, data Page) error {
	return _layoutTmpl.Execute(w, data)
}

// --- HTML Producing Handler ---

// NewHTMLProducingHandler wraps an existing http.Handler that generates raw HTML.
// The output of the given handler is captured and used as the content for a debug page,
// fitting into the standard debug server layout. This is useful for integrating
// http.Handlers that already output HTML directly.
func NewHTMLProducingHandler(source http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// httptest does not import testing package, so we use the recorder directly.
		recorder := httptest.NewRecorder()
		source.ServeHTTP(recorder, r)

		buf := bytes.NewBuffer(nil)
		_layoutTmpl.Execute(buf, Page{
			Title:       "Debug Page", // Generic title, can be overridden by source handler via headers if needed
			NavLinks:    nil,          // NewHTMLProducingHandler doesn't know about other routes for NavLinks
			ContentBody: template.HTML(recorder.Body.String()),
		})

		for key := range recorder.Result().Header {
			w.Header().Set(key, recorder.Result().Header.Get(key))
		}

		w.WriteHeader(recorder.Code)
		buf.WriteTo(w)
	})
}
