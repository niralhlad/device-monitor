package registry

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

/**
Registry stores the set of valid device IDs loaded during application startup.

The registry is read-only after creation and is used to validate that incoming
requests refer to known devices defined in the CSV input file.
*/
type Registry struct {
	deviceIDs map[string]struct{}
}

/**
LoadRegistryFromCSV reads device definitions from a CSV file and builds an in-memory registry.

The function expects a single device ID column. It supports an optional header row,
ignores blank rows, ignores duplicate device IDs, and returns an error when the file
cannot be read or contains no usable device IDs.
*/
func LoadRegistryFromCSV(path string) (*Registry, error) {
	// Reject empty file paths early so startup fails with a clear error.
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("devices csv path cannot be empty")
	}

	// Open the CSV file from disk.
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open devices csv: %w", err)
	}
	defer file.Close()

	// Create a CSV reader for the input file.
	reader := csv.NewReader(file)

	// Allocate the registry map that will hold unique device IDs.
	deviceIDs := make(map[string]struct{})

	// Read each CSV record until the end of the file.
	for {
		// Read the next row from the CSV file.
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read devices csv: %w", readErr)
		}

		// Skip rows that do not contain at least one column.
		if len(record) == 0 {
			continue
		}

		// Read and trim the first column because the challenge CSV uses one device ID column.
		deviceID := strings.TrimSpace(record[0])

		// Skip blank rows.
		if deviceID == "" {
			continue
		}

		// Skip the optional header row.
		if strings.EqualFold(deviceID, "device_id") {
			continue
		}

		// Store the device ID in the registry map.
		deviceIDs[deviceID] = struct{}{}
	}

	// Reject files that did not produce any usable device IDs.
	if len(deviceIDs) == 0 {
		return nil, errors.New("devices csv contains no device IDs")
	}

	// Return the ready-to-use registry.
	return &Registry{
		deviceIDs: deviceIDs,
	}, nil
}

/**
Has reports whether the provided device ID exists in the registry.

This method is used by request handlers and future stores to verify that
incoming telemetry belongs to a known device.
*/
func (r *Registry) Has(deviceID string) bool {
	// Return false when the registry is nil.
	if r == nil {
		return false
	}

	// Look up the device ID in the registry map.
	_, ok := r.deviceIDs[strings.TrimSpace(deviceID)]
	return ok
}

/**
Count returns the total number of unique device IDs in the registry.

This helper is useful during startup logs and unit tests.
*/
func (r *Registry) Count() int {
	// Return zero when the registry is nil.
	if r == nil {
		return 0
	}

	// Return the number of registered device IDs.
	return len(r.deviceIDs)
}

/*
NewForTest creates a registry directly from a list of device IDs.

This helper is used by unit tests that need a small in-memory registry without
creating a temporary CSV file on disk.
*/
func NewForTest(deviceIDs []string) *Registry {
	// Allocate the registry map.
	items := make(map[string]struct{}, len(deviceIDs))

	// Add each provided device ID to the registry map.
	for _, deviceID := range deviceIDs {
		items[strings.TrimSpace(deviceID)] = struct{}{}
	}

	// Return the constructed registry.
	return &Registry{
		deviceIDs: items,
	}
}