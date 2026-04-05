package v1

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/niralhlad/device-monitor/internal/handlers"
	"github.com/niralhlad/device-monitor/internal/registry"
	"github.com/niralhlad/device-monitor/internal/services"
)

/*
TestRegisterDeviceRoutes_RegistersHeartbeatRoute verifies that the version 1 router
registers the heartbeat endpoint successfully.
*/
func TestRegisterDeviceRoutes_RegistersHeartbeatRoute(t *testing.T) {
	// Create a new HTTP mux for the route registration test.
	mux := http.NewServeMux()

	// Create a device handler with a valid service dependency.
	deviceHandler := handlers.NewDeviceHandler(
		services.NewDeviceService(registry.NewForTest([]string{"device-1"})),
	)

	// Register the version 1 device routes.
	registerDeviceRoutes(mux, deviceHandler)

	// Create a valid heartbeat request to the registered route.
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/devices/device-1/heartbeat",
		strings.NewReader(`{"sent_at":"2026-04-05T12:00:10Z"}`),
	)

	// Execute the request against the mux.
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Verify that the route was registered and handled successfully.
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}

/*
TestRegisterDeviceRoutes_PanicsWhenMuxIsNil verifies that route registration fails fast
when the mux dependency is missing.
*/
func TestRegisterDeviceRoutes_PanicsWhenMuxIsNil(t *testing.T) {
	// Recover the expected panic.
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic, got nil")
		}
	}()

	// Attempt to register routes with a nil mux.
	registerDeviceRoutes(nil, handlers.NewDeviceHandler(
		services.NewDeviceService(registry.NewForTest([]string{"device-1"})),
	))
}

/*
TestRegisterDeviceRoutes_PanicsWhenDeviceHandlerIsNil verifies that route registration
fails fast when the device handler dependency is missing.
*/
func TestRegisterDeviceRoutes_PanicsWhenDeviceHandlerIsNil(t *testing.T) {
	// Recover the expected panic.
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic, got nil")
		}
	}()

	// Attempt to register routes with a nil device handler.
	registerDeviceRoutes(http.NewServeMux(), nil)
}

/*
TestRegisterDeviceRoutes_RegistersStatsRoute verifies that the version 1 router
registers the stats endpoint successfully.
*/
func TestRegisterDeviceRoutes_RegistersStatsRoute(t *testing.T) {
	// Create a new HTTP mux for the route registration test.
	mux := http.NewServeMux()

	// Create a device handler with a valid service dependency.
	deviceHandler := handlers.NewDeviceHandler(
		services.NewDeviceService(registry.NewForTest([]string{"device-1"})),
	)

	// Register the version 1 device routes.
	registerDeviceRoutes(mux, deviceHandler)

	// Create a stats request for a known device.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/device-1/stats", nil)

	// Execute the request against the mux.
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Verify that the route was registered and handled successfully.
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}