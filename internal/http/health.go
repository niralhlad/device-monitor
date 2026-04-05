package http

import (
	"net/http"

	"github.com/niralhlad/device-monitor/internal/constants"
	"github.com/niralhlad/device-monitor/internal/handlers"
)

/*
The registerHealthRoutes function registers health check endpoints onto the provided mux.
*/
func registerHealthRoutes(mux *http.ServeMux, healthHandler *handlers.HealthHandler) {
	// Skip registration when the caller does not provide a health handler.
	if mux == nil || healthHandler == nil {
		return
	}

	// Register lightweight health routes used by local runs and container platforms.
	mux.HandleFunc("GET "+constants.HealthBasePath, healthHandler.HandleLive)
}
