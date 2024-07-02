package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	err := app.db.PingContext(r.Context())
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, envelope{}); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}
