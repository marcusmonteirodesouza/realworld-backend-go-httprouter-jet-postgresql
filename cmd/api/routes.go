package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.GET("/healthz", app.healthcheck)

	router.GET("/user", app.authenticate(app.getCurrentUser))
	router.POST("/users", app.registerUser)
	router.POST("/users/login", app.login)
	router.PUT("/user", app.authenticate(app.updateUser))

	router.GET("/profiles/:username", app.authenticateOptional(app.getProfile))
	router.POST("/profiles/:username/follow", app.authenticate(app.followUser))
	router.DELETE("/profiles/:username/follow", app.authenticate(app.unfollowUser))

	return router
}
