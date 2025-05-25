package debughandler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHTMLProducingHandler(t *testing.T) {
	tests := []struct {
		name               string
		sourceContent      string
		sourceStatusCode   int
		sourceHeaders      map[string]string
		expectedBodySubstr string
		expectedStatusCode int
		expectedHeaders    map[string]string
	}{
		{
			name:               "Simple HTML content",
			sourceContent:      "<h1>Hello World</h1><p>This is a test.</p>",
			sourceStatusCode:   http.StatusOK,
			expectedBodySubstr: "<h1>Hello World</h1><p>This is a test.</p>",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Content with different status code",
			sourceContent:      "<h2>Unauthorized</h2>",
			sourceStatusCode:   http.StatusUnauthorized,
			expectedBodySubstr: "<h2>Unauthorized</h2>",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "Content with custom headers",
			sourceContent:      "<p>Custom headers</p>",
			sourceStatusCode:   http.StatusOK,
			sourceHeaders:      map[string]string{"X-Custom-Header": "value1", "Cache-Control": "no-cache"},
			expectedBodySubstr: "<p>Custom headers</p>",
			expectedStatusCode: http.StatusOK,
			expectedHeaders:    map[string]string{"X-Custom-Header": "value1", "Cache-Control": "no-cache"},
		},
		{
			name:               "Empty content",
			sourceContent:      "",
			sourceStatusCode:   http.StatusOK,
			expectedBodySubstr: "", // The layout will still be there
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a source handler that writes the specified content, status, and headers
			sourceHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.sourceHeaders != nil {
					for k, v := range tt.sourceHeaders {
						w.Header().Set(k, v)
					}
				}
				w.WriteHeader(tt.sourceStatusCode)
				_, _ = w.Write([]byte(tt.sourceContent))
			})

			htmlProducingWrapper := NewHTMLProducingHandler(sourceHandler)

			req := httptest.NewRequest("GET", "http://example.com/debug/test", nil)
			rr := httptest.NewRecorder()

			htmlProducingWrapper.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatusCode)
			}

			body := rr.Body.String()
			// Check that the source content is embedded within the layout
			if !strings.Contains(body, tt.expectedBodySubstr) {
				t.Errorf("handler body = %s; want to contain %s", body, tt.expectedBodySubstr)
			}

			// Check if the layout was applied (e.g., by looking for a known part of the layout)
			if !strings.Contains(body, "<title>") || !strings.Contains(body, "YANM Debug") {
				t.Errorf("handler body does not seem to contain the layout: %s", body)
			}

			// Check headers
			if tt.expectedHeaders != nil {
				for k, expectedV := range tt.expectedHeaders {
					if actualV := rr.Header().Get(k); actualV != expectedV {
						t.Errorf("handler returned wrong header %s: got %v want %v", k, actualV, expectedV)
					}
				}
			}
		})
	}
}
