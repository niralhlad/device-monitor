package http

import (
	"log/slog"
	stdhttp "net/http"

	appconfig "github.com/niralhlad/device-monitor/internal/config"
	"github.com/niralhlad/device-monitor/internal/handlers"
	"github.com/niralhlad/device-monitor/internal/http/v1"
)

// Dependencies captures the collaborators required to build the HTTP router.
type Dependencies struct {
	Settings      appconfig.Settings
	Logger        *slog.Logger
	HealthHandler *handlers.HealthHandler
	DeviceHandler *handlers.DeviceHandler
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
	if dependencies.DeviceHandler == nil {
		panic("http NewRouter: nil deviceHandler")
	}

	// Register all root-level routes.
	registerHealthRoutes(mux, dependencies.HealthHandler)

	// Register device monitoring routes.
	v1.RegisterRoutes(mux, dependencies.DeviceHandler)

	// Wrap the router with recovery first so panics are handled safely.
	handler := withRecovery(dependencies.Logger, mux)

	// Wrap the recovered handler with request logging.
	handler = withRequestLogging(dependencies.Logger, handler)

	// Return the final HTTP handler chain.
	return handler
}