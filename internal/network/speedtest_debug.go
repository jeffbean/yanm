package network

import (
	"net/http"
	"text/template"
)

const speedTestDebugHTMLTemplate = `
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
`

var _tempTmpl = template.Must(template.New("speedtest_debug").Parse(speedTestDebugHTMLTemplate))

type page struct {
	s *SpeedTestClient
}

func (p *page) getPageData() ([]*PingResult, []*Performance) {
	p.s.mu.RLock()
	defer p.s.mu.RUnlock()

	var pings []*PingResult
	if len(p.s.lastPingResults) > 0 {
		pings = make([]*PingResult, len(p.s.lastPingResults))
		copy(pings, p.s.lastPingResults)
	}

	var networkTests []*Performance
	if len(p.s.lastNetworkResults) > 0 {
		networkTests = make([]*Performance, len(p.s.lastNetworkResults))
		copy(networkTests, p.s.lastNetworkResults)
	}

	return pings, networkTests
}

func (p *page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pings, networkTests := p.getPageData()
	if err := _tempTmpl.Execute(w, struct {
		Pings        []*PingResult
		NetworkTests []*Performance
		PingCount    int
		NetworkCount int
		MaxHistory   int
	}{
		Pings:        pings,
		NetworkTests: networkTests,
		PingCount:    len(pings),
		NetworkCount: len(networkTests),
		MaxHistory:   maxHistory, // This is the const from speedtest.go
	}); err != nil {
		p.s.logger.ErrorContext(r.Context(), "Failed to execute template", "error", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}

// Debug returns a DebugRoute for the speedtest debug page.

func (s *SpeedTestClient) Debug() http.Handler {
	return &page{s: s}
}
