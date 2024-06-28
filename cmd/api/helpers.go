package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"cloud.google.com/go/logging"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/internal/services"
)

type envelope map[string]interface{}

type errorResponse struct {
	Errors errorResponseErrors `json:"errors"`
}

type errorResponseErrors struct {
	Body []string `json:"body"`
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{msg: msg}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(json); err != nil {
		return err
	}

	return nil
}

func (app *application) writeErrorResponse(w http.ResponseWriter, err error) {
	app.logger.StandardLogger(logging.Error).Print(err.Error())

	var msg string
	var status int

	var alreadyExistsError *services.AlreadyExistsError
	var invalidArgumentError *services.InvalidArgumentError
	var notFoundError *services.NotFoundError

	var forbiddenError *forbiddenError
	var malformedRequest *malformedRequest
	var unauthorizedError *unauthorizedError

	switch {
	case errors.As(err, &alreadyExistsError):
		msg = err.Error()
		status = http.StatusConflict

	case errors.As(err, &forbiddenError):
		msg = "Forbidden"
		status = http.StatusForbidden

	case errors.As(err, &invalidArgumentError):
		msg = err.Error()
		status = http.StatusUnprocessableEntity

	case errors.As(err, &malformedRequest):
		msg = malformedRequest.Error()
		status = http.StatusBadRequest

	case errors.As(err, &notFoundError):
		msg = notFoundError.Error()
		status = http.StatusNotFound

	case errors.As(err, &unauthorizedError):
		msg = "Unauthorized"
		status = http.StatusUnauthorized

	default:
		msg = "Internal server error"
		status = http.StatusInternalServerError
	}

	if err = writeJSON(w, status, errorResponse{
		Errors: errorResponseErrors{
			Body: []string{msg},
		}}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
