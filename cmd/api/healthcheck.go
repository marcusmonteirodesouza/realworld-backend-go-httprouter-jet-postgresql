package main

import "net/http"

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	err := app.db.PingContext(r.Context())
	if err != nil {
		app.writeErrorResponse(w, http.StatusServiceUnavailable, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{}, nil)
	if err != nil {
		app.writeErrorResponse(w, http.StatusInternalServerError, err)
	}
}
