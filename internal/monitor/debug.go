package monitor

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

type monitorPage struct {
	monitor *Network
}

const _monitorPage = `
<h1>Monitor Debug</h1>
<div>
	<h2>Current Limiter States</h2>
	<p>Ping Limiter: {{ .PingLimiter }}</p>
	<p>Network Limiter: {{ .NetworkLimiter }}</p>
</div>
<form action="/debug/monitor/" method="post">
	<button name="action" value="pause-ping">Pause Ping</button>
	<button name="action" value="resume-ping">Resume Ping</button>
	<button name="action" value="pause-network">Pause Network</button>
	<button name="action" value="resume-network">Resume Network</button>
</form>
`

var _monitorPageTemplate = template.Must(template.New("monitor").Parse(_monitorPage))

// NewMonitorDebugPageProvider creates a new debug page provider for the application configuration.
// The handler returned is the raw content-producing handler.
func NewMonitorDebugPageProvider(monitor *Network) http.Handler {
	return &monitorPage{
		monitor: monitor,
	}
}

func (p *monitorPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "text/html") {
			// Serve HTML
			w.Header().Set("Content-Type", "text/html")
			page := struct {
				PingLimiter    string
				NetworkLimiter string
			}{
				PingLimiter:    p.monitor.pingLimiter.Status(),
				NetworkLimiter: p.monitor.networkLimiter.Status(),
			}
			if err := _monitorPageTemplate.Execute(w, page); err != nil {
				http.Error(w, "Failed to execute template", http.StatusInternalServerError)
				return
			}
		} else {
			// Serve JSON
			w.Header().Set("Content-Type", "application/json")
			// Calculate limiter states first
			state := struct {
				PingLimiter    string `json:"ping_limiter"`
				NetworkLimiter string `json:"network_limiter"`
			}{
				PingLimiter:    p.monitor.pingLimiter.Status(),
				NetworkLimiter: p.monitor.networkLimiter.Status(),
			}
			if err := json.NewEncoder(w).Encode(state); err != nil {
				http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
				return
			}
		}

	case http.MethodPost:
		action := r.FormValue("action")
		switch action {
		case "pause-ping":
			p.monitor.PausePing()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ping paused"))
		case "resume-ping":
			p.monitor.ResumePing()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ping resumed"))
		case "pause-network":
			p.monitor.PauseNetwork()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Network paused"))
		case "resume-network":
			p.monitor.ResumeNetwork()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Network resumed"))
		default:
			http.Error(w, "Invalid action", http.StatusBadRequest)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
