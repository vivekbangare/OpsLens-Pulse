package utils

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

type APISuccess struct {
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, errType string, message string, requestID string) {
	WriteJSON(w, status, APIError{
		Error:     errType,
		Message:   message,
		RequestID: requestID,
	})
}

func WriteSuccess(w http.ResponseWriter, status int, data interface{}, message string, requestID string) {
	WriteJSON(w, status, APISuccess{
		Data:      data,
		Message:   message,
		RequestID: requestID,
	})
}
