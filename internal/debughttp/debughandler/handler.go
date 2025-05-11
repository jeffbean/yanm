package debughandler

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"net/http/httptest"
)

var (
	//go:embed layout.html
	_layoutHTMLTemplate string

	_layoutTmpl = template.Must(template.New("layout").Parse(_layoutHTMLTemplate))
)

type page struct {
	// NOTE: RequestContext can be used to pass around values to the template.
	// It is not used by the layout, but can be used by the content.

	// ContentBody is the HTML content to be displayed in the body of the page.
	ContentBody template.HTML
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
		_layoutTmpl.Execute(buf, page{
			ContentBody: template.HTML(recorder.Body.String()),
		})

		for key := range recorder.Result().Header {
			w.Header().Set(key, recorder.Result().Header.Get(key))
		}

		w.WriteHeader(recorder.Code)
		buf.WriteTo(w)
	})
}
