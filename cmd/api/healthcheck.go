package main

import "net/http"

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	err := app.db.PingContext(r.Context())
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, envelope{})
	if err != nil {
		app.writeErrorResponse(w, err)
	}
}
