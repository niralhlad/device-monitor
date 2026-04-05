package http

import (
	"log/slog"
	stdhttp "net/http"

	appconfig "github.com/niralhlad/device-monitor/internal/config"
	"github.com/niralhlad/device-monitor/internal/handlers"
)

// Dependencies captures the collaborators required to build the HTTP router.
type Dependencies struct {
	Settings      appconfig.Settings
	Logger        *slog.Logger
	HealthHandler *handlers.HealthHandler
}

/*
NewRouter builds and returns the root HTTP handler for the service.

The returned handler is the fully registered root router for all
supported endpoints.
*/
func NewRouter(dependencies Dependencies) stdhttp.Handler {
	// Create the root HTTP multiplexer.
	mux := stdhttp.NewServeMux()

	// Validate required dependencies during startup.
	if dependencies.HealthHandler == nil {
		panic("http NewRouter: nil healthHandler")
	}
	if dependencies.Logger == nil {
		panic("http NewRouter: nil logger")
	}

	// Register all root-level routes.
	registerHealthRoutes(mux, dependencies.HealthHandler)

	return mux
}