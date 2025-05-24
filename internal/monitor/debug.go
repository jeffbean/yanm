package monitor

import "net/http"

type monitorPage struct {
	monitor *Network
}

const (
	_monitorDebugPagePath = "/debug/monitor"
)

// NewMonitorDebugPageProvider creates a new debug page provider for the application configuration.
// The handler returned is the raw content-producing handler.
func NewMonitorDebugPageProvider(monitor *Network) http.Handler {
	return &monitorPage{
		monitor: monitor,
	}
}

func (p *monitorPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if GET:
	//  if html content type, render and serve html
	//  else serve json we return json with the monitor state
	// if POST:
	//  if pause ping, pause ping
	//  if resume ping, resume ping
	//  if pause network, pause network
	//  if resume network, resume network
	//  else return error

	w.Write([]byte("Monitor Debug Page"))
}
