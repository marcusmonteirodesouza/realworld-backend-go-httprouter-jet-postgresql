package main

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/model"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/internal/services"
)

type createArticleRequest struct {
	Article createArticleRequestArticle `json:"article"`
}

type createArticleRequestArticle struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Body        string    `json:"body"`
	TagList     *[]string `json:"tagList"`
}

type articleResponse struct {
	Article articleResponseArticle `json:"article"`
}

type articleResponseArticle struct {
	Slug           string                 `json:"slug"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Body           string                 `json:"body"`
	TagList        []string               `json:"tagList"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	Favorited      bool                   `json:"favorited"`
	FavoritesCount int                    `json:"favoritesCount"`
	Author         profileResponseProfile `json:"author"`
}

func newArticleResponse(article model.Article, articleTags []model.ArticleTag, favorited bool, authorProfile services.Profile) articleResponse {
	tagList := make([]string, len(articleTags))
	for i, tag := range articleTags {
		tagList[i] = tag.Name
	}

	return articleResponse{
		Article: articleResponseArticle{
			Slug:        article.Slug,
			Title:       article.Title,
			Description: article.Description,
			Body:        article.Body,
			TagList:     tagList,
			CreatedAt:   *article.CreatedAt,
			UpdatedAt:   *article.UpdatedAt,
			Favorited:   favorited,
			Author:      newProfileResponseProfile(authorProfile),
		},
	}
}

type ListOfTagsResponse struct {
	Tags []string `json:"tags"`
}

func NewListOfTagsResponse(articleTags []model.ArticleTag) ListOfTagsResponse {
	tagList := make([]string, len(articleTags))
	for i, tag := range articleTags {
		tagList[i] = tag.Name
	}

	return ListOfTagsResponse{
		Tags: tagList,
	}
}

func (app *application) createArticle(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var request createArticleRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	ctx := r.Context()

	author := app.contextGetUser(r)

	article, err := app.articlesService.CreateArticle(ctx, services.NewCreateArticle(author.ID, request.Article.Title, request.Article.Description, request.Article.Body, request.Article.TagList))
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	articleTags, err := app.articlesService.ListTags(ctx, services.ListTags{ArticleID: &article.ID})
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	authorProfile, err := app.profilesService.GetProfile(ctx, author.ID, nil)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	articleResponse := newArticleResponse(*article, *articleTags, false, *authorProfile)

	if err = writeJSON(w, http.StatusCreated, articleResponse); err != nil {
		app.writeErrorResponse(w, err)
	}
}

func (app *application) getTags(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	tags, err := app.articlesService.ListTags(ctx, services.NewListTags(nil))
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	listOfTagsResponse := NewListOfTagsResponse(*tags)

	if err = writeJSON(w, http.StatusOK, listOfTagsResponse); err != nil {
		app.writeErrorResponse(w, err)
	}
}
