package v1

import (
	"net/http"

	"github.com/niralhlad/device-monitor/internal/constants"
	"github.com/niralhlad/device-monitor/internal/handlers"
)

/*
registerDeviceRoutes registers all device-related version 1 endpoints.

This function owns only device resource routing and relies on routes.go to
provide the shared version base path.
*/
func registerDeviceRoutes(mux *http.ServeMux, deviceHandler *handlers.DeviceHandler) {
	// Fail fast if required dependencies are missing during application startup.
	if mux == nil {
		panic("v1 registerDeviceRoutes: nil mux")
	}
	if deviceHandler == nil {
		panic("v1 registerDeviceRoutes: nil deviceHandler")
	}

	// Construct the full base path for device-related endpoints by combining the version base path with the device resource path.
	deviceBasePath := constants.APIV1BasePath + constants.DevicesPath 

	// Construct the path for the device ID parameter.
	deviceIDPath := deviceBasePath + constants.DeviceIDParam

	/**
	* @api {post} /api/v1/devices/{device_id}/heartbeat Register device heartbeat
	* @apiName RegisterDeviceHeartbeat
	* @apiGroup Devices
	* @apiDescription Register a heartbeat from a known device.
	*
	* @apiParam {String} device_id ID of the device.
	* @apiBody {String} sent_at Heartbeat timestamp in RFC3339 / ISO 8601 date-time format.
	*
	* @apiBodyExample {json} Request-Body-Example
	* {
	*   "sent_at": "2026-04-04T10:15:30Z"
	* }
	*
	* @apiSuccess (204) NoContent The heartbeat was recorded successfully.
	*
	* @apiErrorExample {json} BadRequest
	* HTTP/1.1 400 Bad Request
	* {
	*   "msg": "invalid request body"
	* }
	*
	* @apiErrorExample {json} DeviceNotFound
	* HTTP/1.1 404 Not Found
	* {
	*   "msg": "device not found"
	* }
	*
	* @apiErrorExample {json} InternalServerError
	* HTTP/1.1 500 Internal Server Error
	* {
	*   "msg": "internal server error"
	* }
	*/
	mux.HandleFunc("POST "+deviceIDPath+"/heartbeat", deviceHandler.HandleHeartbeat)

	/**
	* @api {get} /api/v1/devices/{device_id}/stats Get device statistics
	* @apiName GetDeviceStatistics
	* @apiGroup Devices
	* @apiDescription Return current statistics for a known device.
	*
	* @apiParam {String} device_id ID of the device.
	*
	* @apiSuccess {Number} uptime Device uptime as a percentage.
	* @apiSuccess {String} avg_upload_time Average upload time as a duration string. Currently returns "0s" until upload metric ingestion is implemented.
	*
	* @apiSuccessExample {json} SuccessResponse
	* HTTP/1.1 200 OK
	* {
	*   "uptime": 100,
	*   "avg_upload_time": "0s"
	* }
	*
	* @apiErrorExample {json} DeviceNotFound
	* HTTP/1.1 404 Not Found
	* {
	*   "msg": "device not found"
	* }
	*
	* @apiErrorExample {json} InternalServerError
	* HTTP/1.1 500 Internal Server Error
	* {
	*   "msg": "internal server error"
	* }
	*/
	mux.HandleFunc("GET "+deviceIDPath+"/stats", deviceHandler.HandleGetStats)
}