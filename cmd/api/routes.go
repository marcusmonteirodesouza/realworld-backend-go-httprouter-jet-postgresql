package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/healthz", app.healthcheck)

	router.HandlerFunc(http.MethodPost, "/users", app.registerUser)
	router.HandlerFunc(http.MethodPost, "/users/login", app.login)

	return router
}
