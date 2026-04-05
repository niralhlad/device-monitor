package v1

import (
	"net/http"

	"github.com/niralhlad/device-monitor/internal/handlers"
)

/*
RegisterRoutes registers all version 1 endpoints onto the provided router.

This function is the composition root for version 1 HTTP route registration.
*/
func RegisterRoutes(mux *http.ServeMux, deviceHandler *handlers.DeviceHandler) {
	// Fail fast if required dependencies are missing during application startup.
	if mux == nil {
		panic("v1 RegisterRoutes: nil mux")
	}
	if deviceHandler == nil {
		panic("v1 RegisterRoutes: nil deviceHandler")
	}

	// Register all device-related endpoints under the version 1 base path.
	registerDeviceRoutes(mux, deviceHandler)
}