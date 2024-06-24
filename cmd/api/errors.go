package main

import (
	"net/http"

	"cloud.google.com/go/logging"
)

type errorResponse struct {
	errors errorResponseErrors
}

type errorResponseErrors struct {
	body []string
}

func (app *application) writeErrorResponse(w http.ResponseWriter, status int, err error) {
	app.logger.StandardLogger(logging.Error).Print(err.Error())

	app.writeJSON(w, status, errorResponse{
		errors: errorResponseErrors{
			body: []string{"Internal server error"},
		},
	}, nil)
}
