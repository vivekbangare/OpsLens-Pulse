package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// DecodeJSONStrict safely decodes JSON body into target struct
// - Disallows unknown fields
// - Prevents multiple JSON objects
// - Handles empty body
func DecodeJSONStrict(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	// Ensure only single JSON object
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("multiple JSON objects in body")
	}

	return nil
}
