package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/joho/godotenv"

	"github.com/niralhlad/device-monitor/internal/app"
)

/*
main loads the application dependencies, starts the HTTP server, and handles
graceful shutdown when the process receives an interrupt or termination signal.

The server is configured with the application handler and structured error logger.
When shutdown is requested, the server stops accepting new requests and is given
a bounded amount of time to finish in-flight work before exiting.
*/
func main() {
	// Load environment variables from the .env file for local development.
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	// Load the fully wired application from configuration.
	application, err := app.Load()
	if err != nil {
		log.Printf("failed to load application: %v", err)
		os.Exit(1)
	}

	logger := application.Logger

	// Create the HTTP server using the application settings and root handler.
	server := &http.Server{
		Addr:     application.Settings.Address(),
		Handler:  application.Handler,
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// Listen for interrupt and termination signals to support graceful shutdown.
	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run shutdown handling in the background while the server continues serving traffic.
	go func() {
		<-shutdownContext.Done()

		// Create a bounded shutdown window so the process does not hang indefinitely.
		boundedContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		logger.Warn("shutdown requested")
		if shutdownErr := server.Shutdown(boundedContext); shutdownErr != nil {
			logger.Error("graceful shutdown failed", "error", shutdownErr)
		}
	}()

	logger.Info("server starting", "port", application.Settings.HTTPPort)

	// Start serving traffic and treat unexpected server failures as fatal.
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}