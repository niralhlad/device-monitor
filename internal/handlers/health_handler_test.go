package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

/*
TestNewHealthHandler verifies that NewHealthHandler returns a properly
initialized health handler instance.

The test confirms that the handler is created successfully and that the
provided service name and environment values are stored correctly.
*/
func TestNewHealthHandler(t *testing.T) {
	// Create a new health handler with test metadata.
	h := NewHealthHandler("device-monitor", "test")

	// Ensure the handler instance was created successfully.
	if h == nil {
		t.Fatal("expected handler, got nil")
	}

	// Verify that the service name was stored correctly.
	if h.serviceName != "device-monitor" {
		t.Fatalf("serviceName = %q, want %q", h.serviceName, "device-monitor")
	}

	// Verify that the environment value was stored correctly.
	if h.environment != "test" {
		t.Fatalf("environment = %q, want %q", h.environment, "test")
	}
}

/*
TestHealthHandlerHandleLive verifies that HandleLive returns the expected
successful health response.

The test confirms that the endpoint responds with HTTP 200, sets the JSON
content type, and returns the expected health payload values.
*/
func TestHealthHandlerHandleLive(t *testing.T) {
	// Create a new health handler with test metadata.
	h := NewHealthHandler("device-monitor", "test")

	// Build a test request for the health endpoint.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	// Serve the health request using the live handler.
	h.HandleLive(rr, req)

	// Confirm that the request completed successfully.
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	// Confirm that the response content type is JSON.
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	// Decode the JSON response body into the expected response structure.
	var body HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Confirm that the returned response payload contains the expected values.
	if body.Status != "ok" || body.Service != "device-monitor" || body.Environment != "test" {
		t.Fatalf("unexpected body: %+v", body)
	}
}