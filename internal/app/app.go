package app

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	appconfig "github.com/niralhlad/device-monitor/internal/config"
	"github.com/niralhlad/device-monitor/internal/handlers"
	internalhttp "github.com/niralhlad/device-monitor/internal/http"
)

/*
Application groups the core runtime dependencies created during bootstrap.

Keeping these dependencies together makes the application startup flow
easy to manage and keeps the main entrypoint small.
*/
type Application struct {
	Settings appconfig.Settings
	Logger   *slog.Logger
	Handler  http.Handler
}

/*
Load reads configuration from the process environment and builds the full application.

This function validates startup settings, creates shared dependencies,
registers handlers and routes, and returns the fully constructed application.
*/
func Load() (*Application, error) {
	// Read runtime settings from the process environment.
	settings := appconfig.LoadFromEnv(os.Getenv)

	// Fail early when configuration is invalid.
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	// Create the application logger used across the service.
	logger := newLogger(settings)

	// Create the HTTP handlers used by the API routes.
	healthHandler := handlers.NewHealthHandler(settings.ServiceName, settings.Environment)

	// Register the router and return the top-level HTTP handler.
	router := internalhttp.NewRouter(internalhttp.Dependencies{
		Settings:      settings,
		Logger:        logger,
		HealthHandler: healthHandler,
	})

	// Return the fully constructed application.
	return &Application{
		Settings: settings,
		Logger:   logger,
		Handler:  router,
	}, nil
}

/*
newLogger creates a small structured logger for the application.

The logger supports text or JSON output and provides built-in Info,
Warn, and Error level methods through slog.
*/
func newLogger(settings appconfig.Settings) *slog.Logger {
	// Default to info level.
	level := slog.LevelInfo

	// Apply the configured log level when provided.
	switch strings.ToLower(settings.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	// Configure the logger handler options.
	options := &slog.HandlerOptions{
		Level: level,
	}

	// Choose the output format based on configuration.
	var handler slog.Handler
	if strings.ToLower(settings.LogFormat) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	// Build the logger with stable service metadata attached.
	return slog.New(handler).With(
		"service", settings.ServiceName,
		"environment", settings.Environment,
	)
}