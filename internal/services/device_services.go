package services

import (
	"errors"
	"sync"
	"time"

	"github.com/niralhlad/device-monitor/internal/registry"
)

/*
ErrDeviceNotFound is returned when a request references a device ID that does not exist
in the preloaded device registry.
*/
var ErrDeviceNotFound = errors.New("device not found")

/*
DeviceState represents the in-memory state stored for each device by the service.
FirstMinute and LastMinute track the range of observed heartbeat minutes for uptime calculation.
HasHeartbeat is a simple flag to indicate if any heartbeats have been recorded for the device.
UniqueMinuteCount tracks the total number of unique heartbeat minutes stored for the device.
SeenMinutes is a set of all unique heartbeat minute buckets observed for the device.
*/
type DeviceState struct {
	FirstMinute       int64
	LastMinute        int64
	HasHeartbeat      bool
	UniqueMinuteCount int
	SeenMinutes       map[int64]struct{}
}

/*
DeviceService stores device telemetry state in memory and exposes business operations
used by the device HTTP handler.
*/
type DeviceService struct {
	registry   *registry.Registry
	mu         sync.RWMutex
	devices  map[string]*DeviceState
}

/*
DeviceStats represents the device statistics currently supported by the service.

Uptime is the total number of unique heartbeat minutes recorded for the device
AvgUploadTime is a placeholder for the average upload time metric
*/
type DeviceStats struct {
	Uptime        float64
	AvgUploadTime string
}

/*
NewDeviceService creates a new in-memory device service using the provided registry.

The registry contains the list of valid devices loaded from the CSV file during startup.
*/
func NewDeviceService(deviceRegistry *registry.Registry) *DeviceService {
	// Create an empty in-memory heartbeat store for known devices.
	return &DeviceService{
		registry:   deviceRegistry,
		devices:  make(map[string]*DeviceState),
	}
}

/*
RecordHeartbeat validates the device ID and stores the heartbeat as a UTC minute bucket.

The minute bucket approach keeps heartbeat storage compact and prevents multiple heartbeats
within the same minute from being counted more than once.
*/
func (s *DeviceService) RecordHeartbeat(deviceID string, sentAt time.Time) error {
	// Reject unknown device IDs before storing any data.
	if s.registry == nil || !s.registry.Has(deviceID) {
		return ErrDeviceNotFound
	}

	// Convert the timestamp into a UTC minute bucket.
	minute := sentAt.UTC().Unix() / 60

	// Lock the service before mutating the in-memory heartbeat state.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create the per-device heartbeat map on first use.
	if s.devices[deviceID] == nil {
		s.devices[deviceID] = &DeviceState{
			SeenMinutes: make(map[int64]struct{}),
		}
	}

	device := s.devices[deviceID]

	if _, exists := device.SeenMinutes[minute]; !exists {
		device.SeenMinutes[minute] = struct{}{}
        device.UniqueMinuteCount++

        if !device.HasHeartbeat || minute < device.FirstMinute {
            device.FirstMinute = minute
        }
        if !device.HasHeartbeat || minute > device.LastMinute {
            device.LastMinute = minute
        }

        device.HasHeartbeat = true
	}

	return nil
}

/*
HeartbeatCount returns the number of unique heartbeat minutes stored for the device.

This helper is mainly used by tests in the current commit.
*/
func (s *DeviceService) HeartbeatCount(deviceID string) int {
	// Read-lock the service while inspecting the in-memory state.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return zero when the device has no state yet.
	device := s.devices[deviceID]
	if device == nil {
		return 0
	}

	// Return the current unique heartbeat minute count.
	return device.UniqueMinuteCount
}

/*
GetStats returns the current calculated statistics for the provided device.

At this stage, the service calculates uptime from heartbeat minute buckets and returns
a placeholder upload duration until upload metric ingestion is implemented.
*/
func (s *DeviceService) GetStats(deviceID string) (DeviceStats, error) {
	// Reject unknown device IDs before reading any state.
	if s.registry == nil || !s.registry.Has(deviceID) {
		return DeviceStats{}, ErrDeviceNotFound
	}

	// Read-lock the service while inspecting the in-memory heartbeat state.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read the unique heartbeat minute buckets stored for the device.
	device := s.devices[deviceID]

	// Check if the device has any heartbeat data. If not, return zero uptime and the placeholder upload time.
	if device == nil || !device.HasHeartbeat {
        return DeviceStats{
            Uptime:        0,
            AvgUploadTime: "0s",
        }, nil
    }

	// Handle the edge case of a single heartbeat minute to avoid division by zero and return 100% uptime.
	if device.FirstMinute == device.LastMinute {
		return DeviceStats{
			Uptime:        100,
			AvgUploadTime: "0s",
		}, nil
	}

	// Calculate uptime for the device
	totalMinutes := device.LastMinute - device.FirstMinute + 1
	uptime := (float64(device.UniqueMinuteCount) / float64(totalMinutes)) * 100

	// Return the calculated stats.
	return DeviceStats{
		Uptime:        uptime,
		AvgUploadTime: "0s",
	}, nil
}