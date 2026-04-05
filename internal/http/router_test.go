package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	appconfig "github.com/niralhlad/device-monitor/internal/config"
	"github.com/niralhlad/device-monitor/internal/handlers"
	"github.com/niralhlad/device-monitor/internal/registry"
	"github.com/niralhlad/device-monitor/internal/services"
)

/*
testLogger creates a logger that discards all output during tests.

This helper keeps test output clean while still satisfying dependencies
that require a non-nil logger instance.
*/
func testLogger() *slog.Logger {
	// Return a logger that writes to a discarded output stream.
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

/*
TestNewRouter_RegistersHealthRoute verifies that NewRouter registers
the health endpoint successfully.

The test confirms that a GET request to /health is routed correctly
and returns an HTTP 200 response.
*/
func TestNewRouter_RegistersHealthRoute(t *testing.T) {
	// Build the router with valid dependencies.
	router := NewRouter(Dependencies{
		Settings:      appconfig.DefaultSettings(),
		Logger:        testLogger(),
		HealthHandler: handlers.NewHealthHandler("device-monitor", "test"),
		DeviceHandler: handlers.NewDeviceHandler(
			services.NewDeviceService(registry.NewForTest([]string{"device-1"})),
		),
	})

	// Create a test request for the health endpoint.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	// Serve the request through the router.
	router.ServeHTTP(rr, req)

	// Confirm that the route was registered and handled successfully.
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

/*
TestNewRouter_PanicsWhenHealthHandlerIsNil verifies that NewRouter fails fast
when the health handler dependency is missing.

The test confirms that router construction panics instead of allowing
the application to start in an invalid state.
*/
func TestNewRouter_PanicsWhenHealthHandlerIsNil(t *testing.T) {
	// Recover the expected panic triggered by the missing dependency.
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic, got nil")
		}
	}()

	// Attempt to build the router without a health handler.
	_ = NewRouter(Dependencies{
		Settings: appconfig.DefaultSettings(),
		Logger:   testLogger(),
	})
}

/*
TestNewRouter_PanicsWhenLoggerIsNil verifies that NewRouter fails fast
when the logger dependency is missing.

The test confirms that router construction panics instead of allowing
the application to start with an invalid dependency graph.
*/
func TestNewRouter_PanicsWhenLoggerIsNil(t *testing.T) {
	// Recover the expected panic triggered by the missing dependency.
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic, got nil")
		}
	}()

	// Attempt to build the router without a logger.
	_ = NewRouter(Dependencies{
		Settings:      appconfig.DefaultSettings(),
		HealthHandler: handlers.NewHealthHandler("device-monitor", "test"),
	})
}

/*
TestNewRouter_PanicsWhenDeviceHandlerIsNil verifies that NewRouter fails fast
when the device handler dependency is missing.
*/
func TestNewRouter_PanicsWhenDeviceHandlerIsNil(t *testing.T) {
	// Recover the expected panic triggered by the missing dependency.
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic, got nil")
		}
	}()

	// Attempt to build the router without a device handler.
	_ = NewRouter(Dependencies{
		Settings:      appconfig.DefaultSettings(),
		Logger:        testLogger(),
		HealthHandler: handlers.NewHealthHandler("device-monitor", "test"),
	})
}

/*
TestRegisterHealthRoutes_NoPanicOnNilInputs verifies that registerHealthRoutes
safely ignores nil inputs.

The test confirms that passing a nil mux or nil health handler does not panic
and that no route is registered when the handler is missing.
*/
func TestRegisterHealthRoutes_NoPanicOnNilInputs(t *testing.T) {
	// Confirm that nil inputs are ignored safely.
	registerHealthRoutes(nil, nil)

	// Build a mux and attempt registration with a nil health handler.
	mux := http.NewServeMux()
	registerHealthRoutes(mux, nil)

	// Send a request to the health route, which should remain unregistered.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Confirm that the route was not registered.
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

/*
TestNewRouter_HealthRouteStillWorksWithMiddleware verifies that the router
still serves requests correctly after request logging and recovery middleware
are applied.
*/
func TestNewRouter_HealthRouteStillWorksWithMiddleware(t *testing.T) {
	// Build a router with all required dependencies.
	router := NewRouter(Dependencies{
		Settings: appconfig.DefaultSettings(),
		Logger:   testLogger(),
		HealthHandler: handlers.NewHealthHandler(
			"device-monitor",
			"test",
		),
		DeviceHandler: handlers.NewDeviceHandler(
			services.NewDeviceService(registry.NewForTest([]string{"device-1"})),
		),
	})

	// Create a request to the health endpoint.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Execute the request against the router.
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verify that middleware wrapping did not break routing.
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

/*
TestWithRecovery_ReturnsInternalServerErrorOnPanic verifies that the recovery
middleware catches panics and returns HTTP 500.
*/
func TestWithRecovery_ReturnsInternalServerErrorOnPanic(t *testing.T) {
	// Build a handler that always panics.
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	// Wrap the panic handler with recovery middleware.
	handler := withRecovery(testLogger(), panicHandler)

	// Execute a request against the recovered handler.
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Verify that the panic was converted into HTTP 500.
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}
