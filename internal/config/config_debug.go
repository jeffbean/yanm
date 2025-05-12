package config

import (
	_ "embed"
	"html/template"
	"net/http"

	"gopkg.in/yaml.v3"
)

//go:embed config_debug.html
var configDebugHTMLTemplate string

var configTmpl = template.Must(template.New("config_debug").Parse(configDebugHTMLTemplate))

type configPage struct {
	cfg *Configuration
}

// ServeHTTP handles the request for the configuration debug page.
func (p *configPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	yamlBytes, err := yaml.Marshal(p.cfg)
	if err != nil {
		http.Error(w, "Failed to render configuration", http.StatusInternalServerError)
		return
	}

	data := struct {
		FormattedConfig template.HTML
	}{
		FormattedConfig: template.HTML(string(yamlBytes)),
	}

	if err := configTmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}

// NewConfigDebugPageProvider creates a new debug page provider for the application configuration.
// The handler returned is the raw content-producing handler.
func NewConfigDebugPageProvider(cfg *Configuration) http.Handler {
	return &configPage{
		cfg: cfg,
	}
}
