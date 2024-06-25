package main

import (
	"errors"
	"net/http"

	"cloud.google.com/go/logging"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/internal/services"
)

type errorResponse struct {
	Errors errorResponseErrors `json:"errors"`
}

type errorResponseErrors struct {
	Body []string `json:"body"`
}

func (app *application) writeErrorResponse(w http.ResponseWriter, err error) {
	app.logger.StandardLogger(logging.Error).Print(err.Error())

	var msg string
	var status int

	var alreadyExistsError *services.AlreadyExistsError
	var invalidArgumentError *services.InvalidArgumentError

	var malformedRequest *malformedRequest

	switch {
	case errors.As(err, &alreadyExistsError):
		msg = err.Error()
		status = http.StatusConflict

	case errors.As(err, &invalidArgumentError):
		msg = err.Error()
		status = http.StatusUnprocessableEntity

	case errors.As(err, &malformedRequest):
		msg = malformedRequest.Error()
		status = http.StatusBadRequest

	default:
		msg = "Internal server error"
		status = http.StatusInternalServerError
	}

	err = writeJSON(w, status, errorResponse{
		Errors: errorResponseErrors{
			Body: []string{msg},
		},
	}, nil)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
