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

	router.GET("/articles", app.authenticateOptional(app.listArticles))
	// See https://github.com/gin-gonic/gin/issues/1301
	router.GET("/articles/:slug", func() httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			slug := ps.ByName("slug")

			if slug == "feed" {
				app.authenticate(app.feedArticles)(w, r, ps)
			} else {
				app.authenticateOptional(app.getArticleBySlug)(w, r, ps)
			}
		}
	}())
	router.POST("/articles", app.authenticate(app.createArticle))
	router.POST("/articles/:slug/favorite", app.authenticate(app.favoriteArticle))
	router.DELETE("/articles/:slug/favorite", app.authenticate(app.unfavoriteArticle))

	router.GET("/tags", app.getTags)

	return app.recoverPanic(router)
}
