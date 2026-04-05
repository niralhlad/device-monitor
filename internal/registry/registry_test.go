package registry

import (
	"os"
	"path/filepath"
	"testing"
)

/*
*
TestLoadRegistryFromCSV loads a valid CSV file and verifies that
the registry contains the expected unique device IDs.
*/
func TestLoadRegistryFromCSV(t *testing.T) {
	// Create a temporary CSV file with a header, duplicates, and blank rows.
	path := writeDevicesCSVFile(t, "device_id\nabc-123\nxyz-789\nabc-123\n\n")

	// Load the device registry from the CSV file.
	registry, err := LoadRegistryFromCSV(path)
	if err != nil {
		t.Fatalf("LoadRegistryFromCSV() error = %v", err)
	}

	// Verify that the registry was created.
	if registry == nil {
		t.Fatal("expected registry, got nil")
	}

	// Verify that duplicate IDs were de-duplicated.
	if got := registry.Count(); got != 2 {
		t.Fatalf("Count() = %d, want %d", got, 2)
	}

	// Verify that known device IDs are present.
	if !registry.Has("abc-123") {
		t.Fatal("expected registry to contain abc-123")
	}
	if !registry.Has("xyz-789") {
		t.Fatal("expected registry to contain xyz-789")
	}

	// Verify that unknown device IDs are rejected.
	if registry.Has("missing-device") {
		t.Fatal("expected registry to reject missing-device")
	}
}

/*
*
TestLoadRegistryFromCSV_ReturnsErrorForMissingFile verifies that startup fails
when the configured CSV file path does not exist.
*/
func TestLoadRegistryFromCSV_ReturnsErrorForMissingFile(t *testing.T) {
	// Attempt to load a registry from a file that does not exist.
	registry, err := LoadRegistryFromCSV(filepath.Join(t.TempDir(), "missing.csv"))

	// Verify that an error is returned.
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify that no registry is returned on error.
	if registry != nil {
		t.Fatalf("expected nil registry, got %+v", registry)
	}
}

/*
*
TestLoadRegistryFromCSV_ReturnsErrorForEmptyPath verifies that empty paths
are rejected before any file access is attempted.
*/
func TestLoadRegistryFromCSV_ReturnsErrorForEmptyPath(t *testing.T) {
	// Attempt to load a registry using an empty file path.
	registry, err := LoadRegistryFromCSV("   ")

	// Verify that an error is returned.
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify that no registry is returned on error.
	if registry != nil {
		t.Fatalf("expected nil registry, got %+v", registry)
	}
}

/*
*
TestLoadRegistryFromCSV_ReturnsErrorForEmptyCSV verifies that files without any
usable device IDs are rejected during startup.
*/
func TestLoadRegistryFromCSV_ReturnsErrorForEmptyCSV(t *testing.T) {
	// Create a temporary CSV file that contains only the header row.
	path := writeDevicesCSVFile(t, "device_id\n")

	// Attempt to load the registry from the empty CSV file.
	registry, err := LoadRegistryFromCSV(path)

	// Verify that an error is returned.
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify that no registry is returned on error.
	if registry != nil {
		t.Fatalf("expected nil registry, got %+v", registry)
	}
}

/*
*
TestRegistryHas_ReturnsFalseForNilReceiver verifies that the registry helper
is safe to call even when the receiver is nil.
*/
func TestRegistryHas_ReturnsFalseForNilReceiver(t *testing.T) {
	// Declare a nil registry pointer.
	var registry *Registry

	// Verify that nil registries safely return false.
	if registry.Has("abc-123") {
		t.Fatal("expected nil registry to return false")
	}
}

/*
*
TestRegistryCount_ReturnsZeroForNilReceiver verifies that the registry helper
is safe to call even when the receiver is nil.
*/
func TestRegistryCount_ReturnsZeroForNilReceiver(t *testing.T) {
	// Declare a nil registry pointer.
	var registry *Registry

	// Verify that nil registries safely return zero.
	if got := registry.Count(); got != 0 {
		t.Fatalf("Count() = %d, want %d", got, 0)
	}
}

/*
*
writeDevicesCSVFile creates a temporary CSV file for registry tests.

The helper writes the provided contents to disk and returns the generated path.
*/
func writeDevicesCSVFile(t *testing.T, contents string) string {
	// Mark this helper as a test helper.
	t.Helper()

	// Build a temporary file path inside the test directory.
	path := filepath.Join(t.TempDir(), "devices.csv")

	// Write the provided CSV contents to disk.
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Return the generated file path.
	return path
}
