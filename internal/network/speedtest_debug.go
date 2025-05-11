package network

import (
	"html/template"
	"net/http"
)

const speedTestDebugHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Speed Test Debug</title>
    <style>
        body { font-family: sans-serif; margin: 20px; }
        table { border-collapse: collapse; width: 100%; margin-bottom: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        h2 { margin-top: 30px; }
    </style>
</head>
<body>
    <h1>Speed Test Results</h1>

    <h2>Last {{.PingCount}} Ping Tests (Max {{.MaxHistory}})</h2>
    {{if .Pings}}
    <table>
        <tr>
            <th>Timestamp</th>
            <th>Target Server</th>
            <th>Latency</th>
        </tr>
        {{range .Pings}}
        <tr>
            <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
            <td>{{.TargetName}}</td>
            <td>{{.Latency}}</td>
        </tr>
        {{end}}
    </table>
    {{else}}
    <p>No ping test results yet.</p>
    {{end}}

    <h2>Last {{.NetworkCount}} Network Speed Tests (Max {{.MaxHistory}})</h2>
    {{if .NetworkTests}}
    <table>
        <tr>
            <th>Timestamp</th>
            <th>Target Server</th>
            <th>Download (Mbps)</th>
            <th>Upload (Mbps)</th>
            <th>Ping Latency</th>
        </tr>
        {{range .NetworkTests}}
        <tr>
            <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
            <td>{{.TargetName}}</td>
            <td>{{printf "%.2f" .DownloadSpeedMbps}}</td>
            <td>{{printf "%.2f" .UploadSpeedMbps}}</td>
            <td>{{.PingLatency}}</td>
        </tr>
        {{end}}
    </table>
    {{else}}
    <p>No network speed test results yet.</p>
    {{end}}
</body>
</html>
`

// MetricsHandler returns an http.Handler that serves an HTML page with recent speed test results.
func (s *SpeedTestClient) MetricsHandler() http.Handler {
	tmpl, err := template.New("speedTestDebug").Parse(speedTestDebugHTMLTemplate)
	if err != nil {
		s.logger.Error("Failed to parse speedtest debug HTML template", "error", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal server error: could not load debug template", http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		// Make copies of the slices to avoid holding the lock during template execution
		var pings []*PingResult
		if len(s.lastPingResults) > 0 {
			pings = make([]*PingResult, len(s.lastPingResults))
			copy(pings, s.lastPingResults)
		}

		var networkTests []*NetworkPerformance
		if len(s.lastNetworkResults) > 0 {
			networkTests = make([]*NetworkPerformance, len(s.lastNetworkResults))
			copy(networkTests, s.lastNetworkResults)
		}
		s.mu.RUnlock()

		data := struct {
			Pings        []*PingResult
			NetworkTests []*NetworkPerformance
			PingCount    int
			NetworkCount int
			MaxHistory   int
		}{
			Pings:        pings,
			NetworkTests: networkTests,
			PingCount:    len(pings),
			NetworkCount: len(networkTests),
			MaxHistory:   maxHistory,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			s.logger.ErrorContext(r.Context(), "Failed to execute speedtest debug HTML template", "error", err)
			// If headers haven't been sent, try to send an error
			if w.Header().Get("Content-Type") == "" {
				http.Error(w, "Internal server error: could not render debug page", http.StatusInternalServerError)
			}
		}
	})
}
