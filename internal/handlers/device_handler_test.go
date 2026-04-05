package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/niralhlad/device-monitor/internal/services"
	"github.com/niralhlad/device-monitor/internal/registry"
)

/*
TestNewDeviceHandler creates the device handler successfully with a valid service dependency.
*/
func TestNewDeviceHandler(t *testing.T) {
	// Create a device service for the handler dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))

	// Create the device handler.
	handler := NewDeviceHandler(deviceService)

	// Verify that the handler was created successfully.
	if handler == nil {
		t.Fatal("expected handler, got nil")
	}
}

/*
TestHandleHeartbeat records a heartbeat successfully for a known device.
*/
func TestHandleHeartbeat(t *testing.T) {
	// Create a device handler with a valid service dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))
	handler := NewDeviceHandler(deviceService)

	// Build a valid heartbeat request body.
	body := `{"sent_at":"2026-04-05T12:00:10Z"}`

	// Create the HTTP request and set the route path value.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/device-1/heartbeat", strings.NewReader(body))
	req.SetPathValue("device_id", "device-1")

	// Execute the request against the handler.
	rr := httptest.NewRecorder()
	handler.HandleHeartbeat(rr, req)

	// Verify that the request completed successfully.
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}

	// Verify that the heartbeat was stored in memory.
	if got := deviceService.HeartbeatCount("device-1"); got != 1 {
		t.Fatalf("HeartbeatCount() = %d, want %d", got, 1)
	}
}

/*
TestHandleHeartbeat_ReturnsBadRequestForInvalidJSON verifies that malformed request bodies
are rejected with HTTP 400.
*/
func TestHandleHeartbeat_ReturnsBadRequestForInvalidJSON(t *testing.T) {
	// Create a device handler with a valid service dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))
	handler := NewDeviceHandler(deviceService)

	// Create a malformed JSON request body.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/device-1/heartbeat", strings.NewReader(`{"sent_at":`))
	req.SetPathValue("device_id", "device-1")

	// Execute the request against the handler.
	rr := httptest.NewRecorder()
	handler.HandleHeartbeat(rr, req)

	// Verify that the request is rejected as invalid input.
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

/*
TestHandleHeartbeat_ReturnsBadRequestForMissingSentAt verifies that requests without
a valid sent_at value are rejected with HTTP 400.
*/
func TestHandleHeartbeat_ReturnsBadRequestForMissingSentAt(t *testing.T) {
	// Create a device handler with a valid service dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))
	handler := NewDeviceHandler(deviceService)

	// Create a request body missing the sent_at field.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/device-1/heartbeat", strings.NewReader(`{}`))
	req.SetPathValue("device_id", "device-1")

	// Execute the request against the handler.
	rr := httptest.NewRecorder()
	handler.HandleHeartbeat(rr, req)

	// Verify that the request is rejected as invalid input.
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

/*
TestHandleHeartbeat_ReturnsNotFoundForUnknownDevice verifies that the handler returns HTTP 404
when the request targets a device not present in the registry.
*/
func TestHandleHeartbeat_ReturnsNotFoundForUnknownDevice(t *testing.T) {
	// Create a device handler with a valid service dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))
	handler := NewDeviceHandler(deviceService)

	// Build a valid heartbeat request for an unknown device.
	body := `{"sent_at":"2026-04-05T12:00:10Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/missing-device/heartbeat", strings.NewReader(body))
	req.SetPathValue("device_id", "missing-device")

	// Execute the request against the handler.
	rr := httptest.NewRecorder()
	handler.HandleHeartbeat(rr, req)

	// Verify that the unknown device is rejected.
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

/*
TestHandleHeartbeat_DeduplicatesSameMinute verifies that multiple requests in the same minute
still result in a single unique stored heartbeat minute.
*/
func TestHandleHeartbeat_DeduplicatesSameMinute(t *testing.T) {
	// Create a device handler with a valid service dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))
	handler := NewDeviceHandler(deviceService)

	// Record two heartbeats for the same device within the same minute.
	for _, body := range []string{
		`{"sent_at":"2026-04-05T12:00:10Z"}`,
		`{"sent_at":"2026-04-05T12:00:50Z"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/device-1/heartbeat", strings.NewReader(body))
		req.SetPathValue("device_id", "device-1")

		rr := httptest.NewRecorder()
		handler.HandleHeartbeat(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
		}
	}

	// Verify that only one unique heartbeat minute was stored.
	if got := deviceService.HeartbeatCount("device-1"); got != 1 {
		t.Fatalf("HeartbeatCount() = %d, want %d", got, 1)
	}
}

/*
TestHandleGetStats_ReturnsStatsForKnownDevice verifies that the stats endpoint returns
the currently calculated uptime for a known device.
*/
func TestHandleGetStats_ReturnsStatsForKnownDevice(t *testing.T) {
	// Create a device handler with a valid service dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))
	handler := NewDeviceHandler(deviceService)

	// Store one heartbeat so the device has measurable uptime.
	if err := deviceService.RecordHeartbeat("device-1", time.Date(2026, 4, 5, 12, 0, 10, 0, time.UTC)); err != nil {
		t.Fatalf("RecordHeartbeat() error = %v", err)
	}

	// Build the stats request for the known device.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/device-1/stats", nil)
	req.SetPathValue("device_id", "device-1")

	// Execute the request against the handler.
	rr := httptest.NewRecorder()
	handler.HandleGetStats(rr, req)

	// Verify the response status code.
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	// Verify the response body contains the expected fields.
	body := rr.Body.String()
	if !strings.Contains(body, `"uptime":100`) {
		t.Fatalf("body = %q, want uptime field", body)
	}
	if !strings.Contains(body, `"avg_upload_time":"0s"`) {
		t.Fatalf("body = %q, want avg_upload_time field", body)
	}
}

/*
TestHandleGetStats_ReturnsZeroStatsForKnownDeviceWithoutHeartbeats verifies that a known device
without heartbeat data still returns a successful stats response.
*/
func TestHandleGetStats_ReturnsZeroStatsForKnownDeviceWithoutHeartbeats(t *testing.T) {
	// Create a device handler with a valid service dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))
	handler := NewDeviceHandler(deviceService)

	// Build the stats request for the known device.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/device-1/stats", nil)
	req.SetPathValue("device_id", "device-1")

	// Execute the request against the handler.
	rr := httptest.NewRecorder()
	handler.HandleGetStats(rr, req)

	// Verify the response status code.
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	// Verify the response body contains zero uptime.
	body := rr.Body.String()
	if !strings.Contains(body, `"uptime":0`) {
		t.Fatalf("body = %q, want zero uptime", body)
	}
}

/*
TestHandleGetStats_ReturnsNotFoundForUnknownDevice verifies that the stats endpoint
rejects unknown device IDs with HTTP 404.
*/
func TestHandleGetStats_ReturnsNotFoundForUnknownDevice(t *testing.T) {
	// Create a device handler with a valid service dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))
	handler := NewDeviceHandler(deviceService)

	// Build the stats request for an unknown device.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/missing-device/stats", nil)
	req.SetPathValue("device_id", "missing-device")

	// Execute the request against the handler.
	rr := httptest.NewRecorder()
	handler.HandleGetStats(rr, req)

	// Verify that the unknown device is rejected.
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

/*
TestHandleGetStats_ReturnsHundredForConsecutiveHeartbeatMinutes verifies that two
consecutive heartbeat minutes produce 100 percent uptime.
*/
func TestHandleGetStats_ReturnsHundredForConsecutiveHeartbeatMinutes(t *testing.T) {
	// Create a device handler with a valid service dependency.
	deviceService := services.NewDeviceService(registry.NewForTest([]string{"device-1"}))
	handler := NewDeviceHandler(deviceService)

	// Store two consecutive heartbeat minutes for the device.
	for _, ts := range []time.Time{
		time.Date(2026, 4, 1, 10, 30, 1, 0, time.UTC),
		time.Date(2026, 4, 1, 10, 31, 1, 0, time.UTC),
	} {
		if err := deviceService.RecordHeartbeat("device-1", ts); err != nil {
			t.Fatalf("RecordHeartbeat() error = %v", err)
		}
	}

	// Build the stats request for the known device.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/device-1/stats", nil)
	req.SetPathValue("device_id", "device-1")

	// Execute the request against the handler.
	rr := httptest.NewRecorder()
	handler.HandleGetStats(rr, req)

	// Verify the response status code.
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	// Verify the response body contains 100 uptime.
	body := rr.Body.String()
	if !strings.Contains(body, `"uptime":100`) {
		t.Fatalf("body = %q, want uptime 100", body)
	}
}