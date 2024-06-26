package main

import (
	"fmt"
	"net/http"
	"strings"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection:", "close")
				app.writeErrorResponse(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getToken(r)

		if token == "" {
			app.writeErrorResponse(w, &unauthorizedError{msg: "Invalid or missing authentication token"})
			return
		}

		app.serveHTTPAuthenticated(next, w, r, token)
	})
}

func (app *application) authenticateOptional(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getToken(r)

		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		app.serveHTTPAuthenticated(next, w, r, token)
	})
}

func (app *application) serveHTTPAuthenticated(next http.Handler, w http.ResponseWriter, r *http.Request, token string) {
	ctx := r.Context()

	user, err := app.usersService.GetUserByToken(ctx, token)
	if err != nil {
		app.writeErrorResponse(w, &unauthorizedError{msg: err.Error()})
		return
	}

	r = app.contextSetUser(r, user)
	r = app.contextSetToken(r, token)

	next.ServeHTTP(w, r)
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
