package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

/*
TestWriteJSON verifies that WriteJSON writes the expected status code,
content type, and JSON response body.

The test confirms that the helper serializes the provided payload
correctly and marks the response as application/json.
*/
func TestWriteJSON(t *testing.T) {
	// Create a response recorder and a simple JSON payload.
	rr := httptest.NewRecorder()
	payload := map[string]string{"status": "ok"}

	// Write the JSON response using the helper.
	WriteJSON(rr, http.StatusCreated, payload)

	// Confirm that the expected status code was written.
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}

	// Confirm that the response content type is JSON.
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	// Decode the JSON body and confirm that the payload was written correctly.
	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if got["status"] != "ok" {
		t.Fatalf("status field = %q, want %q", got["status"], "ok")
	}
}

/*
TestWriteNoContent verifies that WriteNoContent writes an HTTP 204 response
with no response body.

The test confirms that the helper is suitable for successful operations
that intentionally return no content.
*/
func TestWriteNoContent(t *testing.T) {
	// Create a response recorder for the test.
	rr := httptest.NewRecorder()

	// Write the no-content response.
	WriteNoContent(rr)

	// Confirm that the expected 204 status was returned.
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}

	// Confirm that the response body is empty.
	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rr.Body.String())
	}
}

/*
TestWriteMessage verifies that WriteMessage writes the expected status code
and standard message response body.

The test confirms that the helper wraps the provided message in the shared
Message response structure.
*/
func TestWriteMessage(t *testing.T) {
	// Create a response recorder for the test.
	rr := httptest.NewRecorder()

	// Write the message response with a custom status code.
	WriteMessage(rr, http.StatusAccepted, "queued")

	// Confirm that the expected status code was returned.
	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}

	// Decode the response body and verify the message content.
	var got Message
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if got.Msg != "queued" {
		t.Fatalf("msg = %q, want %q", got.Msg, "queued")
	}
}

/*
TestWriteBadRequest verifies that WriteBadRequest writes an HTTP 400 response.

The test confirms that the helper returns the correct status code for
client-side request validation failures.
*/
func TestWriteBadRequest(t *testing.T) {
	// Create a response recorder for the test.
	rr := httptest.NewRecorder()

	// Write the bad request response.
	WriteBadRequest(rr, "invalid body")

	// Confirm that the expected 400 status was returned.
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

/*
TestWriteNotFound verifies that WriteNotFound writes an HTTP 404 response.

The test confirms that the helper returns the correct status code for
missing resources.
*/
func TestWriteNotFound(t *testing.T) {
	// Create a response recorder for the test.
	rr := httptest.NewRecorder()

	// Write the not-found response.
	WriteNotFound(rr, "missing")

	// Confirm that the expected 404 status was returned.
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

/*
TestWriteInternalServerError verifies that WriteInternalServerError writes
an HTTP 500 response with the standard internal server error message.

The test confirms that the helper returns the correct status code and
stable error payload for unexpected failures.
*/
func TestWriteInternalServerError(t *testing.T) {
	// Create a response recorder for the test.
	rr := httptest.NewRecorder()

	// Write the internal server error response.
	WriteInternalServerError(rr)

	// Confirm that the expected 500 status was returned.
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	// Decode the response body and verify the standard error message.
	var got Message
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if got.Msg != "internal server error" {
		t.Fatalf("msg = %q, want %q", got.Msg, "internal server error")
	}
}