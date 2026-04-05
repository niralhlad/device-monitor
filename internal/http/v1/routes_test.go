package v1

import (
	"net/http"
	"testing"

	"github.com/niralhlad/device-monitor/internal/handlers"
	"github.com/niralhlad/device-monitor/internal/registry"
	"github.com/niralhlad/device-monitor/internal/services"
)

/*
TestRegisterRoutes_PanicsWhenMuxIsNil verifies that version 1 route registration fails fast
when the mux is missing.
*/
func TestRegisterRoutes_PanicsWhenMuxIsNil(t *testing.T) {
	// Recover the expected panic.
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic, got nil")
		}
	}()

	// Attempt to register routes with a nil mux.
	RegisterRoutes(nil, handlers.NewDeviceHandler(
		services.NewDeviceService(registry.NewForTest([]string{"device-1"})),
	))
}

/*
TestRegisterRoutes_PanicsWhenDeviceHandlerIsNil verifies that version 1 route registration
fails fast when the device handler is missing.
*/
func TestRegisterRoutes_PanicsWhenDeviceHandlerIsNil(t *testing.T) {
	// Recover the expected panic.
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic, got nil")
		}
	}()

	// Attempt to register routes with a nil device handler.
	RegisterRoutes(http.NewServeMux(), nil)
}
