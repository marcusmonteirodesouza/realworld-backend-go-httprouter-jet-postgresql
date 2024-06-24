package main

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any,
	headers http.Header) error {
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(json); err != nil {
		return err
	}

	return nil
}
