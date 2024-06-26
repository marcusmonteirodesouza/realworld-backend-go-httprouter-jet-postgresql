package main

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (app *application) authenticate(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		token := getToken(r)

		if token == "" {
			app.writeErrorResponse(w, &unauthorizedError{msg: "Invalid or missing authentication token"})
			return
		}

		app.serveHTTPAuthenticated(h, w, r, ps, token)
	}
}

func (app *application) authenticateOptional(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		token := getToken(r)

		if token == "" {
			h(w, r, ps)
			return
		}

		app.serveHTTPAuthenticated(h, w, r, ps, token)
	}
}

func (app *application) serveHTTPAuthenticated(h httprouter.Handle, w http.ResponseWriter, r *http.Request, ps httprouter.Params, token string) {
	ctx := r.Context()

	user, err := app.usersService.GetUserByToken(ctx, token)
	if err != nil {
		app.writeErrorResponse(w, &unauthorizedError{msg: err.Error()})
		return
	}

	r = app.contextSetUser(r, user)
	r = app.contextSetToken(r, token)

	h(w, r, ps)
}

func getToken(r *http.Request) string {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		authorizationHeader = r.Header.Get("authorization")
	}

	if authorizationHeader == "" {
		return authorizationHeader
	}

	authenticationScheme := "Token"

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != authenticationScheme {
		return ""
	}

	return headerParts[1]
}
