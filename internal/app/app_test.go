package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

/*
TestLoad_ReturnsApplicationWithDefaults verifies that Load builds a valid
application when no environment overrides are provided.

The test confirms that default settings are applied and that the core
runtime dependencies, including the logger and HTTP handler, are initialized.
*/
func TestLoad_ReturnsApplicationWithDefaults(t *testing.T) {
	// Clear relevant environment variables so defaults are used.
	clearAppEnv(t)

	// Load the application using default configuration values.
	application, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Ensure the application bundle was created successfully.
	if application == nil {
		t.Fatal("expected application, got nil")
	}

	// Ensure the shared logger was initialized.
	if application.Logger == nil {
		t.Fatal("expected logger, got nil")
	}

	// Ensure the root HTTP handler was initialized.
	if application.Handler == nil {
		t.Fatal("expected handler, got nil")
	}

	// Confirm that the default HTTP port was applied.
	if application.Settings.HTTPPort != "6733" {
		t.Fatalf("HTTPPort = %q, want %q", application.Settings.HTTPPort, "6733")
	}
}

/*
TestLoad_ReturnsValidationErrorForInvalidPort verifies that Load fails fast
when the configured HTTP port is invalid.

The test ensures that application startup does not continue when configuration
validation fails.
*/
func TestLoad_ReturnsValidationErrorForInvalidPort(t *testing.T) {
	// Clear the environment first to avoid interference from prior values.
	clearAppEnv(t)

	// Set an invalid HTTP port to trigger validation failure.
	t.Setenv("HTTP_PORT", "0")

	// Attempt to load the application with invalid configuration.
	application, err := Load()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	// Ensure no partially built application is returned on failure.
	if application != nil {
		t.Fatalf("expected nil application on error, got %+v", application)
	}
}

/*
TestLoad_BuildsHealthRouteUsingEnvironmentValues verifies that the health
endpoint is built using the configured environment values.

The test confirms that the handler responds successfully and includes the
service name and environment values loaded from environment variables.
*/
func TestLoad_BuildsHealthRouteUsingEnvironmentValues(t *testing.T) {
	// Clear the environment so only test-specific values are used.
	clearAppEnv(t)

	// Set environment values that should appear in the health response.
	t.Setenv("SERVICE_NAME", "fleet-metrics")
	t.Setenv("ENVIRONMENT", "test")

	// Load the application with the test environment values.
	application, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Build a test request for the health endpoint.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	// Serve the request using the application's root handler.
	application.Handler.ServeHTTP(rr, req)

	// Confirm that the request completed successfully.
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	// Confirm that the response was returned as JSON.
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	// Confirm that the response body contains the configured service metadata.
	if body := rr.Body.String(); !strings.Contains(body, "fleet-metrics") || !strings.Contains(body, "test") {
		t.Fatalf("body = %q, want service and environment values", body)
	}
}

/*
clearAppEnv removes the application-specific environment variables used by the tests.

This helper ensures that each test starts from a clean and predictable
environment before setting its own values.
*/
func clearAppEnv(t *testing.T) {
	t.Helper()

	// Remove all application-related environment variables used in these tests.
	for _, key := range []string{"SERVICE_NAME", "ENVIRONMENT", "HTTP_PORT", "LOG_LEVEL", "LOG_FORMAT"} {
		_ = os.Unsetenv(key)
	}
}