package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/healthz", app.healthcheck)

	router.HandlerFunc(http.MethodGet, "/user", app.authenticate(app.getCurrentUser))
	router.HandlerFunc(http.MethodPost, "/users/login", app.login)
	router.HandlerFunc(http.MethodPost, "/users", app.registerUser)

	return app.recoverPanic(router)
}
