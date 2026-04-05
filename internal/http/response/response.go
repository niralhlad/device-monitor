package response

import (
	"encoding/json"
	"net/http"
)

/*
Message defines the standard JSON message response used by the service.

This structure keeps error responses consistent across handlers by exposing
a single "msg" field in the response body.
*/
type Message struct {
	Msg string `json:"msg"`
}

/*
WriteJSON writes the provided payload as a JSON HTTP response.

The function sets the response content type to application/json, writes the
supplied HTTP status code, and encodes the payload into the response body.

If JSON encoding fails, the function falls back to a plain internal server
error response because the intended payload could not be serialized safely.
*/
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	// Mark the response body as JSON.
	w.Header().Set("Content-Type", "application/json")

	// Write the HTTP status before writing the body.
	w.WriteHeader(status)

	// Encode the payload into JSON.
	// Fall back to a plain HTTP error if JSON serialization failed.
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

/*
WriteNoContent writes an HTTP 204 No Content response.

This helper is used for successful operations that intentionally return
no response body.
*/
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

/*
WriteMessage writes a JSON message response with the provided HTTP status code.

This helper wraps the given message in the shared Message response structure
so that all message-based responses follow the same format.
*/
func WriteMessage(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, Message{Msg: message})
}

/*
WriteBadRequest writes a standard HTTP 400 Bad Request response.

This helper should be used when the client sends invalid input, malformed
JSON, or otherwise incorrect request data.
*/
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteMessage(w, http.StatusBadRequest, message)
}

/*
WriteNotFound writes a standard HTTP 404 Not Found response.

This helper should be used when the requested resource cannot be found.
*/
func WriteNotFound(w http.ResponseWriter, message string) {
	WriteMessage(w, http.StatusNotFound, message)
}

/*
WriteInternalServerError writes a standard HTTP 500 Internal Server Error response.

This helper returns the stable internal server error payload used throughout
the service for unexpected failures.
*/
func WriteInternalServerError(w http.ResponseWriter) {
	WriteMessage(w, http.StatusInternalServerError, "internal server error")
}
