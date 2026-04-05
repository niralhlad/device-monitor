package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/niralhlad/device-monitor/internal/http/response"
	"github.com/niralhlad/device-monitor/internal/services"
)

/*
DeviceHandler serves HTTP endpoints related to device telemetry ingestion.
*/
type DeviceHandler struct {
	deviceService *services.DeviceService
}

/*
heartbeatRequest defines the request body expected by the heartbeat endpoint.
*/
type heartbeatRequest struct {
	SentAt time.Time `json:"sent_at"`
}

/*
deviceUploadStatsRequest defines the request body expected by the upload stats endpoint.
*/
type deviceUploadStatsRequest struct {
	SentAt     *string `json:"sent_at"`
	UploadTime int64   `json:"upload_time"`
}

/*
deviceStatsResponse defines the JSON response returned by the device stats endpoint.
*/
type deviceStatsResponse struct {
	Uptime        float64 `json:"uptime"`
	AvgUploadTime string  `json:"avg_upload_time"`
}

/*
NewDeviceHandler creates a new device handler with the provided device service dependency.
*/
func NewDeviceHandler(deviceService *services.DeviceService) *DeviceHandler {
	// Return a ready-to-use device handler.
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

/*
HandleHeartbeat validates the request payload and records a heartbeat for the given device.

The endpoint stores the heartbeat as a UTC minute bucket and returns HTTP 204 on success.
*/
func (h *DeviceHandler) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	// Read the device ID path parameter from the request URL.
	deviceID := r.PathValue("device_id")

	// Decode the JSON request body into the heartbeat request shape.
	var req heartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteBadRequest(w, "invalid request body")
		return
	}

	// Reject missing or zero-value timestamps.
	if req.SentAt.IsZero() {
		response.WriteBadRequest(w, "sent_at is required")
		return
	}

	// Record the heartbeat in the device service.
	if err := h.deviceService.RecordHeartbeat(deviceID, req.SentAt); err != nil {
		if errors.Is(err, services.ErrDeviceNotFound) {
			response.WriteNotFound(w, "device not found")
			return
		}

		response.WriteInternalServerError(w)
		return
	}

	// Return a successful no-content response when the heartbeat is recorded.
	response.WriteNoContent(w)
}

/*
HandleGetStats returns the current calculated statistics for the requested device.

At this stage, the endpoint returns uptime derived from heartbeat data and a placeholder
average upload duration until upload metric ingestion is implemented.
*/
func (h *DeviceHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	// Read the device ID path parameter from the request URL.
	deviceID := r.PathValue("device_id")

	// Read the current calculated stats from the device service.
	stats, err := h.deviceService.GetStats(deviceID)
	if err != nil {
		if errors.Is(err, services.ErrDeviceNotFound) {
			response.WriteNotFound(w, "device not found")
			return
		}

		response.WriteInternalServerError(w)
		return
	}

	// Return the stats response as JSON.
	response.WriteJSON(w, http.StatusOK, deviceStatsResponse{
		Uptime:        stats.Uptime,
		AvgUploadTime: stats.AvgUploadTime,
	})
}

/*
HandlePostStats validates the request payload and stores upload telemetry for the given device.

The endpoint aggregates upload durations in memory and returns HTTP 204 on success.
*/
func (h *DeviceHandler) HandlePostStats(w http.ResponseWriter, r *http.Request) {
	// Read the device ID path parameter from the request URL.
	deviceID := r.PathValue("device_id")

	// Decode the JSON request body into the upload stats request shape.
	var req deviceUploadStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		println("POST /stats decode error:", err.Error())
		response.WriteBadRequest(w, "invalid request body")
		return
	}

	if req.SentAt == nil || strings.TrimSpace(*req.SentAt) == "" {
		response.WriteBadRequest(w, "sent_at is required")
		return
	}

	// Reject negative upload durations.
	if req.UploadTime < 0 {
		println("POST /stats negative upload_time")
		response.WriteBadRequest(w, "upload_time must be non-negative")
		return
	}

	// Store the upload stat in the device service.
	if err := h.deviceService.RecordUploadStat(deviceID, req.UploadTime); err != nil {
		if errors.Is(err, services.ErrDeviceNotFound) {
			response.WriteNotFound(w, "device not found")
			return
		}

		response.WriteInternalServerError(w)
		return
	}

	// Return a successful no-content response when the upload stat is recorded.
	response.WriteNoContent(w)
}
