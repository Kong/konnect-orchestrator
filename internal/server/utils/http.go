package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// ReadJSON reads JSON from the request body
func ReadJSON(r *http.Request, v interface{}) error {
	// Check if the Content-Type is application/json
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return errors.New("Content-Type is not application/json")
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// Unmarshal the JSON
	return json.Unmarshal(body, v)
}

// WriteJSON writes JSON to the response
func WriteJSON(w http.ResponseWriter, status int, v interface{}) error {
	// Marshal the JSON
	js, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	return err
}

// WriteError writes an error to the response
func WriteError(w http.ResponseWriter, status int, message string) error {
	return WriteJSON(w, status, map[string]string{
		"error": message,
	})
}
