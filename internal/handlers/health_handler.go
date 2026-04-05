package handlers

import (
	"net/http"

	"github.com/niralhlad/device-monitor/internal/http/response"
)

/*
HealthResponse defines the JSON response body returned by the health endpoints.

This payload provides a small and stable response that can be used by
load balancers, orchestrators, monitoring systems, and manual checks.
*/
type HealthResponse struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
}

/*
HealthHandler serves the operational endpoints used for liveness and readiness checks.

The handler stores static service metadata so both health endpoints can return
consistent response payloads without repeating the same values in each request.
*/
type HealthHandler struct {
	serviceName string
	environment string
}

/*
NewHealthHandler creates and returns a health handler with the provided service metadata.

The returned handler can be reused across all health-related routes and provides
stable service information in each response payload.
*/
func NewHealthHandler(serviceName, environment string) *HealthHandler {
	// Return a ready-to-use health handler instance.
	return &HealthHandler{
		serviceName: serviceName,
		environment: environment,
	}
}

/*
HandleLive handles the liveness endpoint for the service.

This endpoint confirms that the process is running and able to respond to
HTTP requests. It is intended for basic operational checks.
*/
func (h *HealthHandler) HandleLive(w http.ResponseWriter, _ *http.Request) {
	// Return a simple success payload indicating the process is alive.
	response.WriteJSON(w, http.StatusOK, HealthResponse{
		Status:      "ok",
		Service:     h.serviceName,
		Environment: h.environment,
	})
}
