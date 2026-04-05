package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/niralhlad/device-monitor/internal/http/response"
)

/*
statusRecorder wraps http.ResponseWriter so middleware can capture
the final HTTP status code written by the handler chain.
*/
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

/*
WriteHeader captures the response status code before delegating
to the underlying ResponseWriter.
*/
func (r *statusRecorder) WriteHeader(statusCode int) {
	// Store the written HTTP status code for logging.
	r.statusCode = statusCode

	// Forward the status code to the wrapped ResponseWriter.
	r.ResponseWriter.WriteHeader(statusCode)
}

/*
withRequestLogging logs one structured entry for every HTTP request.

The middleware records the request method, request path, response status,
remote address, and total request duration.
*/
func withRequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	// Return the next handler unchanged when logging dependencies are missing.
	if logger == nil || next == nil {
		return next
	}

	// Wrap the next handler with request logging.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the request start time for duration measurement.
		start := time.Now()

		// Wrap the response writer so the status code can be observed.
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Execute the downstream handler chain.
		next.ServeHTTP(recorder, r)

		// Log the completed request with useful debugging fields.
		logger.Info(
			"http request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", recorder.statusCode,
			"remote_addr", r.RemoteAddr,
			"duration", time.Since(start).String(),
		)
	})
}

/*
withRecovery catches panics from downstream handlers, logs the failure,
and returns HTTP 500 instead of crashing the server.
*/
func withRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	// Return the next handler unchanged when dependencies are missing.
	if logger == nil || next == nil {
		return next
	}

	// Wrap the next handler with panic recovery.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Recover from any panic raised by downstream handlers.
		defer func() {
			if recovered := recover(); recovered != nil {
				// Log the recovered panic for debugging.
				logger.Error(
					"http request panic recovered",
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
					"panic", fmt.Sprint(recovered),
				)

				// Return a generic internal server error response.
				response.WriteInternalServerError(w)
			}
		}()

		// Execute the downstream handler chain.
		next.ServeHTTP(w, r)
	})
}