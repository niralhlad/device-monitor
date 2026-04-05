package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	devicesCSVPath := writeDevicesCSVFile(t, "device_id\nabc-123\nxyz-789\n")

	// Point the application to the temporary devices CSV file.
	t.Setenv("DEVICES_CSV_PATH", devicesCSVPath)

	// Attempt to load the application.
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

	// Create a temporary devices CSV file so only the port is invalid.
	devicesCSVPath := writeDevicesCSVFile(t, "device_id\nabc-123\n")

	// Set invalid port and valid devices CSV path.
	t.Setenv("HTTP_PORT", "0")
	t.Setenv("DEVICES_CSV_PATH", devicesCSVPath)

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

	// Create a temporary devices CSV file for startup.
	devicesCSVPath := writeDevicesCSVFile(t, "device_id\nabc-123\nxyz-789\n")

	// Configure the application environment values for the test.
	t.Setenv("SERVICE_NAME", "fleet-metrics")
	t.Setenv("ENVIRONMENT", "test")
	t.Setenv("DEVICES_CSV_PATH", devicesCSVPath)

	// Load the application.
	application, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Send a request to the health route.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	application.Handler.ServeHTTP(rr, req)

	// Verify the HTTP response.
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}
	if body := rr.Body.String(); !strings.Contains(body, "fleet-metrics") || !strings.Contains(body, "test") {
		t.Fatalf("body = %q, want service and environment values", body)
	}
}

/**
TestLoad_ReturnsErrorWhenDevicesCSVIsMissing verifies that bootstrap fails
when the configured devices CSV file does not exist.
*/
func TestLoad_ReturnsErrorWhenDevicesCSVIsMissing(t *testing.T) {
	// Clear environment variables used by the application.
	clearAppEnv(t)

	// Point the application at a missing devices CSV file.
	t.Setenv("DEVICES_CSV_PATH", filepath.Join(t.TempDir(), "missing.csv"))

	// Attempt to load the application.
	application, err := Load()

	// Verify that startup fails.
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if application != nil {
		t.Fatalf("expected nil application on error, got %+v", application)
	}
}

/*
clearAppEnv removes the application-specific environment variables used by the tests.

This helper ensures that each test starts from a clean and predictable
environment before setting its own values.
*/
func clearAppEnv(t *testing.T) {
	t.Helper()

	// Remove all application-specific environment variables.
	for _, key := range []string{
		"SERVICE_NAME",
		"ENVIRONMENT",
		"HTTP_PORT",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"DEVICES_CSV_PATH",
	} {
		_ = os.Unsetenv(key)
	}
}

/**
writeDevicesCSVFile creates a temporary devices CSV file for application bootstrap tests.

The helper writes the provided contents to disk and returns the generated file path.
*/
func writeDevicesCSVFile(t *testing.T, contents string) string {
	// Mark this helper as a test helper.
	t.Helper()

	// Build a temporary file path inside the test directory.
	path := filepath.Join(t.TempDir(), "devices.csv")

	// Write the provided CSV contents to disk.
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Return the generated file path.
	return path
}