package config

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
)

//go:embed config_debug.html
var _configDebugHTMLTemplate string

// HandleConfigDump creates an http.HandlerFunc that dynamically writes the application
// configuration as JSON or an HTML page based on the request's Accept header.
func (c *Configuration) HandleConfigDump(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptHeader := r.Header.Get("Accept")

		// Marshal the config to JSON once
		configJSON, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to marshal configuration", "error", err)
			http.Error(w, "Failed to marshal configuration", http.StatusInternalServerError)
			return
		}

		if strings.Contains(acceptHeader, "application/json") {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write(configJSON); err != nil {
				logger.ErrorContext(r.Context(), "Failed to write configuration JSON", "error", err)
				return
			}
			return
		}
		// Default to HTML
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl, err := template.New("configDebug").Parse(_configDebugHTMLTemplate)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to parse config debug HTML template", "error", err)
			http.Error(w, "Internal server error: could not load config debug template", http.StatusInternalServerError)
			return
		}

		data := struct {
			ConfigJSON string
		}{
			ConfigJSON: string(configJSON),
		}

		if err := tmpl.Execute(w, data); err != nil {
			logger.ErrorContext(r.Context(), "Failed to execute config debug HTML template", "error", err)
			http.Error(w, "Internal server error: could not execute config debug template", http.StatusInternalServerError)
		}
	})
}
