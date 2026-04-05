package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

/*
Settings defines the runtime configuration used by the API service.

This structure contains the core configuration values required to start
and run the application in different environments.
*/
type Settings struct {
	ServiceName string
	Environment string
	HTTPPort    string
	LogLevel    string
	LogFormat   string
}

/*
DefaultSettings returns the default configuration values for the service.

These defaults provide a safe baseline for local development and allow
the service to start with minimal environment setup.
*/
func DefaultSettings() Settings {
	// Return the default application configuration values.
	return Settings{
		ServiceName: "device-monitor",
		Environment: "local",
		HTTPPort:    "6733",
		LogLevel:    "info",
		LogFormat:   "json",
	}
}

/*
LoadFromEnv reads configuration values from environment lookups and merges them with defaults.

The function starts from the default settings and replaces individual fields
only when a non-empty environment value is provided.
*/
func LoadFromEnv(lookup func(string) string) Settings {
	// Use a nil-safe lookup function so tests can inject simple maps.
	if lookup == nil {
		lookup = func(string) string { return "" }
	}

	// Start from the shared defaults and override fields one by one.
	settings := DefaultSettings()

	// Override each setting when a non-empty environment value is available.
	settings.ServiceName = FallbackString(lookup("SERVICE_NAME"), settings.ServiceName)
	settings.Environment = FallbackString(lookup("ENVIRONMENT"), settings.Environment)
	settings.HTTPPort = FallbackString(lookup("HTTP_PORT"), settings.HTTPPort)
	settings.LogLevel = FallbackString(lookup("LOG_LEVEL"), settings.LogLevel)
	settings.LogFormat = FallbackString(lookup("LOG_FORMAT"), settings.LogFormat)

	// Return the merged runtime settings.
	return settings
}

/*
Validate checks whether the configuration contains valid runtime values.

The function currently verifies that the configured HTTP port is present
and falls within the valid TCP port range.
*/
func (s Settings) Validate() error {
	// Ensure the port is set before attempting numeric validation.
	if strings.TrimSpace(s.HTTPPort) == "" {
		return errors.New("HTTP_PORT cannot be empty")
	}

	// Convert the configured port to a number and verify it is in range.
	port, err := strconv.Atoi(s.HTTPPort)
	if err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("HTTP_PORT must be between 1 and 65535")
	}

	// Return nil when all configuration checks pass.
	return nil
}

/*
FallbackString returns the fallback value when the input is empty after trimming whitespace.

This helper is used to merge environment-provided values with defaults
while treating blank strings as missing values.
*/
func FallbackString(value, fallback string) string {
	// Return the fallback when the provided value is empty or whitespace only.
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	// Return the cleaned input value when it is present.
	return strings.TrimSpace(value)
}

/*
Address constructs the full network address for the HTTP server based on the configured port.

This helper method joins the configured TCP port into the standard address format used by net/http.
*/
func (s Settings) Address() string {
	// Join the configured TCP port into the standard address format.
	return ":" + s.HTTPPort
}