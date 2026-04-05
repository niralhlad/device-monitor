package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

/*
TestStatusRecorder_WriteHeader verifies that the status recorder captures
the HTTP status code written by downstream handlers.
*/
func TestStatusRecorder_WriteHeader(t *testing.T) {
	// Create a recorder and wrap it with the status recorder.
	base := httptest.NewRecorder()
	recorder := &statusRecorder{
		ResponseWriter: base,
		statusCode:     http.StatusOK,
	}

	// Write a custom status code through the recorder.
	recorder.WriteHeader(http.StatusCreated)

	// Verify that the recorder captured the status code.
	if recorder.statusCode != http.StatusCreated {
		t.Fatalf("statusCode = %d, want %d", recorder.statusCode, http.StatusCreated)
	}
}