package services

import (
	"testing"
	"time"

	"github.com/niralhlad/device-monitor/internal/registry"
)

/*
TestNewDeviceService creates the service successfully with a valid device registry.
*/
func TestNewDeviceService(t *testing.T) {
	// Create a registry with known device IDs.
	deviceRegistry := registry.NewForTest([]string{"device-1"})

	// Create the device service.
	service := NewDeviceService(deviceRegistry)

	// Verify that the service was created successfully.
	if service == nil {
		t.Fatal("expected service, got nil")
	}
}

/*
TestRecordHeartbeat stores a heartbeat for a known device successfully.
*/
func TestRecordHeartbeat(t *testing.T) {
	// Create a device service with one valid device.
	service := NewDeviceService(registry.NewForTest([]string{"device-1"}))

	// Record a heartbeat for the valid device.
	err := service.RecordHeartbeat("device-1", time.Date(2026, 4, 5, 12, 0, 10, 0, time.UTC))
	if err != nil {
		t.Fatalf("RecordHeartbeat() error = %v", err)
	}

	// Verify that the device state was created and updated correctly.
	device := service.devices["device-1"]
	if device == nil {
		t.Fatal("expected device state, got nil")
	}

	if device.UniqueMinuteCount != 1 {
		t.Fatalf("UniqueMinuteCount = %d, want %d", device.UniqueMinuteCount, 1)
	}

	if !device.HasHeartbeat {
		t.Fatal("expected HasHeartbeat = true")
	}

	if len(device.SeenMinutes) != 1 {
		t.Fatalf("len(SeenMinutes) = %d, want %d", len(device.SeenMinutes), 1)
	}
}

/*
TestRecordHeartbeat_DeduplicatesSameMinute verifies that multiple heartbeats in the same
minute are stored as a single unique heartbeat bucket.
*/
func TestRecordHeartbeat_DeduplicatesSameMinute(t *testing.T) {
	// Create a device service with one valid device.
	service := NewDeviceService(registry.NewForTest([]string{"device-1"}))

	// Record two heartbeats that fall into the same UTC minute.
	first := time.Date(2026, 4, 5, 12, 0, 10, 0, time.UTC)
	second := time.Date(2026, 4, 5, 12, 0, 50, 0, time.UTC)

	if err := service.RecordHeartbeat("device-1", first); err != nil {
		t.Fatalf("first RecordHeartbeat() error = %v", err)
	}
	if err := service.RecordHeartbeat("device-1", second); err != nil {
		t.Fatalf("second RecordHeartbeat() error = %v", err)
	}

	// Verify that only one unique minute bucket was stored.
	device := service.devices["device-1"]
	if device == nil {
		t.Fatal("expected device state, got nil")
	}

	if device.UniqueMinuteCount != 1 {
		t.Fatalf("UniqueMinuteCount = %d, want %d", device.UniqueMinuteCount, 1)
	}

	if len(device.SeenMinutes) != 1 {
		t.Fatalf("len(SeenMinutes) = %d, want %d", len(device.SeenMinutes), 1)
	}
}

/*
TestRecordHeartbeat_ReturnsErrorForUnknownDevice verifies that heartbeats for devices not
present in the registry are rejected.
*/
func TestRecordHeartbeat_ReturnsErrorForUnknownDevice(t *testing.T) {
	// Create a device service with one valid device.
	service := NewDeviceService(registry.NewForTest([]string{"device-1"}))

	// Attempt to record a heartbeat for an unknown device.
	err := service.RecordHeartbeat("missing-device", time.Now())

	// Verify that the service rejects unknown device IDs.
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrDeviceNotFound {
		t.Fatalf("error = %v, want %v", err, ErrDeviceNotFound)
	}
}

/*
TestGetStats_ReturnsNotFoundForUnknownDevice verifies that stats cannot be read
for a device not present in the registry.
*/
func TestGetStats_ReturnsNotFoundForUnknownDevice(t *testing.T) {
	// Create a device service with one valid device.
	service := NewDeviceService(registry.NewForTest([]string{"device-1"}))

	// Attempt to read stats for an unknown device.
	_, err := service.GetStats("missing-device")

	// Verify that the service rejects unknown device IDs.
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrDeviceNotFound {
		t.Fatalf("error = %v, want %v", err, ErrDeviceNotFound)
	}
}

/*
TestGetStats_ReturnsZeroUptimeWithoutHeartbeats verifies that a known device with
no heartbeat data returns zero uptime.
*/
func TestGetStats_ReturnsZeroUptimeWithoutHeartbeats(t *testing.T) {
	// Create a device service with one valid device.
	service := NewDeviceService(registry.NewForTest([]string{"device-1"}))

	// Read stats before recording any heartbeat.
	stats, err := service.GetStats("device-1")
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	// Verify the default stats.
	if stats.Uptime != 0 {
		t.Fatalf("Uptime = %v, want %v", stats.Uptime, 0.0)
	}
	if stats.AvgUploadTime != "0s" {
		t.Fatalf("AvgUploadTime = %q, want %q", stats.AvgUploadTime, "0s")
	}
}

/*
TestGetStats_ReturnsHundredForSingleHeartbeatMinute verifies that one unique heartbeat
minute produces 100 percent uptime.
*/
func TestGetStats_ReturnsHundredForSingleHeartbeatMinute(t *testing.T) {
	// Create a device service with one valid device.
	service := NewDeviceService(registry.NewForTest([]string{"device-1"}))

	// Record one heartbeat.
	err := service.RecordHeartbeat("device-1", time.Date(2026, 4, 5, 12, 0, 10, 0, time.UTC))
	if err != nil {
		t.Fatalf("RecordHeartbeat() error = %v", err)
	}

	// Read stats after recording one heartbeat minute.
	stats, err := service.GetStats("device-1")
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	// Verify the calculated uptime.
	if stats.Uptime != 100 {
		t.Fatalf("Uptime = %v, want %v", stats.Uptime, 100.0)
	}
	if stats.AvgUploadTime != "0s" {
		t.Fatalf("AvgUploadTime = %q, want %q", stats.AvgUploadTime, "0s")
	}
}

/*
TestRecordHeartbeat_UpdatesFirstAndLastMinute verifies that the service tracks
the earliest and latest observed heartbeat minute for a device.
*/
func TestRecordHeartbeat_UpdatesFirstAndLastMinute(t *testing.T) {
	// Create a device service with one valid device.
	service := NewDeviceService(registry.NewForTest([]string{"device-1"}))

	// Record heartbeats in different minutes.
	first := time.Date(2026, 4, 5, 12, 0, 10, 0, time.UTC)
	last := time.Date(2026, 4, 5, 12, 2, 10, 0, time.UTC)

	if err := service.RecordHeartbeat("device-1", last); err != nil {
		t.Fatalf("first RecordHeartbeat() error = %v", err)
	}
	if err := service.RecordHeartbeat("device-1", first); err != nil {
		t.Fatalf("second RecordHeartbeat() error = %v", err)
	}

	// Read the internal device state.
	device := service.devices["device-1"]
	if device == nil {
		t.Fatal("expected device state, got nil")
	}

	// Verify the tracked first and last minute values.
	wantFirst := first.UTC().Unix() / 60
	wantLast := last.UTC().Unix() / 60

	if device.FirstMinute != wantFirst {
		t.Fatalf("FirstMinute = %d, want %d", device.FirstMinute, wantFirst)
	}
	if device.LastMinute != wantLast {
		t.Fatalf("LastMinute = %d, want %d", device.LastMinute, wantLast)
	}
}

/*
TestGetStats_TwoConsecutiveHeartbeatMinutesReturnHundred verifies that two consecutive
heartbeat minutes produce 100 percent uptime.
*/
func TestGetStats_TwoConsecutiveHeartbeatMinutesReturnHundred(t *testing.T) {
	// Create a device service with one valid device.
	service := NewDeviceService(registry.NewForTest([]string{"device-1"}))

	// Record two consecutive heartbeat minutes.
	if err := service.RecordHeartbeat("device-1", time.Date(2026, 4, 1, 10, 30, 1, 0, time.UTC)); err != nil {
		t.Fatalf("first RecordHeartbeat() error = %v", err)
	}
	if err := service.RecordHeartbeat("device-1", time.Date(2026, 4, 1, 10, 31, 1, 0, time.UTC)); err != nil {
		t.Fatalf("second RecordHeartbeat() error = %v", err)
	}

	// Read the calculated device stats.
	stats, err := service.GetStats("device-1")
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	// Verify the uptime calculation.
	if stats.Uptime != 100 {
		t.Fatalf("Uptime = %v, want %v", stats.Uptime, 100.0)
	}
}