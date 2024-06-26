package main

import (
	"context"
	"net/http"

	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/model"
)

type contextKey string

const userContextKey = contextKey("user")

const tokenContextKey = contextKey("token")

func (app *application) contextSetUser(r *http.Request, user *model.Users) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextSetToken(r *http.Request, token string) *http.Request {
	ctx := context.WithValue(r.Context(), tokenContextKey, token)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *model.Users {
	user, ok := r.Context().Value(userContextKey).(*model.Users)
	if !ok {
		panic("Missing user value in request context")
	}

	return user
}

func (app *application) contextGetToken(r *http.Request) string {
	token, ok := r.Context().Value(tokenContextKey).(string)
	if !ok {
		panic("Missing token value in request context")
	}

	return token
}
