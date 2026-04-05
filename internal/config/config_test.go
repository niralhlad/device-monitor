package config

import "testing"

/*
TestDefaultSettings verifies that DefaultSettings returns the expected
baseline configuration values for the application.

The test confirms that the default service name, environment, HTTP port,
log level, and log format are initialized as intended.
*/
func TestDefaultSettings(t *testing.T) {
	// Load the default application settings.
	s := DefaultSettings()

	// Verify the default service name.
	if s.ServiceName != "device-monitor" {
		t.Fatalf("ServiceName = %q, want %q", s.ServiceName, "device-monitor")
	}

	// Verify the default environment.
	if s.Environment != "local" {
		t.Fatalf("Environment = %q, want %q", s.Environment, "local")
	}

	// Verify the default HTTP port.
	if s.HTTPPort != "6733" {
		t.Fatalf("HTTPPort = %q, want %q", s.HTTPPort, "6733")
	}

	// Verify the default log level.
	if s.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want %q", s.LogLevel, "info")
	}

	// Verify the default log format.
	if s.LogFormat != "json" {
		t.Fatalf("LogFormat = %q, want %q", s.LogFormat, "json")
	}
	if s.DevicesCSVPath != "data/devices.csv" {
		t.Fatalf("DevicesCSVPath = %q, want %q", s.DevicesCSVPath, "data/devices.csv")
	}
}

/*
TestLoadFromEnv_UsesDefaultsWhenLookupIsNil verifies that LoadFromEnv falls back
to the default settings when no lookup function is provided.

This ensures that configuration loading remains safe even when the caller
does not provide an environment source.
*/
func TestLoadFromEnv_UsesDefaultsWhenLookupIsNil(t *testing.T) {
	// Load settings without providing an environment lookup function.
	s := LoadFromEnv(nil)

	// Compare the result to the defaults.
	defaults := DefaultSettings()

	// Confirm that the loaded settings match the defaults exactly.
	if s != defaults {
		t.Fatalf("settings = %+v, want %+v", s, defaults)
	}
}

/*
TestLoadFromEnv_OverridesOnlyProvidedValues verifies that LoadFromEnv replaces
default values only for environment variables that are explicitly provided.

The test confirms that all provided values override the defaults correctly.
*/
func TestLoadFromEnv_OverridesOnlyProvidedValues(t *testing.T) {
	// Build a test environment with explicit overrides.
	env := map[string]string{
		"SERVICE_NAME":     "fleet-api",
		"ENVIRONMENT":      "production",
		"HTTP_PORT":        "8080",
		"LOG_LEVEL":        "debug",
		"LOG_FORMAT":       "text",
		"DEVICES_CSV_PATH": "/tmp/devices.csv",
	}

	// Load settings from the test environment.
	s := LoadFromEnv(func(key string) string { return env[key] })

	// Verify that all provided values override the defaults.
	if s.ServiceName != "fleet-api" {
		t.Fatalf("ServiceName = %q, want %q", s.ServiceName, "fleet-api")
	}
	if s.Environment != "production" {
		t.Fatalf("Environment = %q, want %q", s.Environment, "production")
	}
	if s.HTTPPort != "8080" {
		t.Fatalf("HTTPPort = %q, want %q", s.HTTPPort, "8080")
	}
	if s.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want %q", s.LogLevel, "debug")
	}
	if s.LogFormat != "text" {
		t.Fatalf("LogFormat = %q, want %q", s.LogFormat, "text")
	}
	if s.DevicesCSVPath != "/tmp/devices.csv" {
		t.Fatalf("DevicesCSVPath = %q, want %q", s.DevicesCSVPath, "/tmp/devices.csv")
	}
}

/*
TestLoadFromEnv_IgnoresBlankValues verifies that blank or whitespace-only
environment values do not overwrite valid defaults.

This ensures that missing or malformed environment input does not result
in empty configuration values.
*/
func TestLoadFromEnv_IgnoresBlankValues(t *testing.T) {
	// Build a test environment with blank values.
	env := map[string]string{
		"SERVICE_NAME":     "   ",
		"HTTP_PORT":        "",
		"DEVICES_CSV_PATH": "   ",
	}

	// Load settings from the test environment.
	s := LoadFromEnv(func(key string) string { return env[key] })

	// Compare the result to the defaults for the blank fields.
	defaults := DefaultSettings()

	// Confirm that the default service name is preserved.
	if s.ServiceName != defaults.ServiceName {
		t.Fatalf("ServiceName = %q, want default %q", s.ServiceName, defaults.ServiceName)
	}

	// Confirm that the default HTTP port is preserved.
	if s.HTTPPort != defaults.HTTPPort {
		t.Fatalf("HTTPPort = %q, want default %q", s.HTTPPort, defaults.HTTPPort)
	}
	if s.DevicesCSVPath != defaults.DevicesCSVPath {
		t.Fatalf("DevicesCSVPath = %q, want default %q", s.DevicesCSVPath, defaults.DevicesCSVPath)
	}
}

/*
TestSettingsValidate verifies that Settings.Validate accepts valid ports
and rejects invalid HTTP port values.

The test covers valid, empty, whitespace, non-numeric, zero, negative,
and out-of-range port values.
*/
func TestSettingsValidate(t *testing.T) {
	// Define validation scenarios for the HTTP port.
	tests := []struct {
		name           string
		port           string
		devicesCSVPath string
		wantErr        bool
	}{
		{name: "valid default settings", port: "6733", devicesCSVPath: "data/devices.csv", wantErr: false},
		{name: "empty port", port: "", devicesCSVPath: "data/devices.csv", wantErr: true},
		{name: "whitespace port", port: "   ", devicesCSVPath: "data/devices.csv", wantErr: true},
		{name: "non numeric port", port: "abc", devicesCSVPath: "data/devices.csv", wantErr: true},
		{name: "zero port", port: "0", devicesCSVPath: "data/devices.csv", wantErr: true},
		{name: "negative port", port: "-1", devicesCSVPath: "data/devices.csv", wantErr: true},
		{name: "too large port", port: "70000", devicesCSVPath: "data/devices.csv", wantErr: true},
		{name: "empty csv path", port: "6733", devicesCSVPath: "", wantErr: true},
	}

	// Execute each validation test case independently.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start from default settings and override the port under test.
			s := DefaultSettings()
			s.HTTPPort = tt.port
			s.DevicesCSVPath = tt.devicesCSVPath

			// Validate the settings.
			err := s.Validate()

			// Confirm that invalid ports fail validation.
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}

			// Confirm that valid ports pass validation.
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

/*
TestFallbackString verifies that FallbackString returns the trimmed input value
when it is present and returns the fallback when the input is blank.

This ensures that configuration parsing handles whitespace and empty values safely.
*/
func TestFallbackString(t *testing.T) {
	// Confirm that non-empty input is trimmed and returned.
	if got := FallbackString(" value ", "fallback"); got != "value" {
		t.Fatalf("got %q, want %q", got, "value")
	}

	// Confirm that blank input returns the fallback value.
	if got := FallbackString("   ", "fallback"); got != "fallback" {
		t.Fatalf("got %q, want %q", got, "fallback")
	}
}

/*
TestSettingsAddress verifies that Address returns the expected listen address
for the configured HTTP port.

This ensures that the helper produces the standard net/http address format.
*/
func TestSettingsAddress(t *testing.T) {
	// Start from defaults and override the port for the test case.
	s := DefaultSettings()
	s.HTTPPort = "9000"

	// Confirm that the address is formatted correctly.
	if got := s.Address(); got != ":9000" {
		t.Fatalf("Address() = %q, want %q", got, ":9000")
	}
}
