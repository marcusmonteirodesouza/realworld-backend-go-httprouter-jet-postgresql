package main

import "net/http"

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	headers := http.Header{}

	err := app.writeJSON(w, http.StatusOK, envelope{}, headers)
	if err != nil {
		app.writeErrorResponse(w, http.StatusInternalServerError, err)
	}
}
