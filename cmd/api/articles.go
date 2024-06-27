package main

import (
	"context"
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

func newArticleResponse(article model.Article, articleTags []model.ArticleTag, favorited bool, favoritesCount int, authorProfile services.Profile) articleResponse {
	tagList := make([]string, len(articleTags))
	for i, tag := range articleTags {
		tagList[i] = tag.Name
	}

	return articleResponse{
		Article: articleResponseArticle{
			Slug:           article.Slug,
			Title:          article.Title,
			Description:    article.Description,
			Body:           article.Body,
			TagList:        tagList,
			CreatedAt:      *article.CreatedAt,
			UpdatedAt:      *article.UpdatedAt,
			Favorited:      favorited,
			FavoritesCount: favoritesCount,
			Author:         newProfileResponseProfile(authorProfile),
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

	user := app.contextGetUser(r)

	article, err := app.articlesService.CreateArticle(ctx, services.NewCreateArticle(user.ID, request.Article.Title, request.Article.Description, request.Article.Body, request.Article.TagList))
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	articleResponse, err := app.makeArticleResponse(ctx, user, *article)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	if err = writeJSON(w, http.StatusCreated, articleResponse); err != nil {
		app.writeErrorResponse(w, err)
	}
}

func (app *application) getArticleBySlug(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	articleSlug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, articleSlug)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	articleResponse, err := app.makeArticleResponse(ctx, user, *article)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	if err = writeJSON(w, http.StatusCreated, articleResponse); err != nil {
		app.writeErrorResponse(w, err)
	}
}

func (app *application) favoriteArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	articleSlug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, articleSlug)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	if err = app.articlesService.FavoriteArticle(ctx, user.ID, article.ID); err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	article, err = app.articlesService.GetArticleBySlug(ctx, article.Slug)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	articleResponse, err := app.makeArticleResponse(ctx, user, *article)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	if err = writeJSON(w, http.StatusCreated, articleResponse); err != nil {
		app.writeErrorResponse(w, err)
	}
}

func (app *application) unfavoriteArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	articleSlug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, articleSlug)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	if err = app.articlesService.UnfavoriteArticle(ctx, user.ID, article.ID); err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	article, err = app.articlesService.GetArticleBySlug(ctx, article.Slug)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	articleResponse, err := app.makeArticleResponse(ctx, user, *article)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

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

func (app *application) makeArticleResponse(ctx context.Context, user *model.Users, article model.Article) (*articleResponse, error) {
	articleTags, err := app.articlesService.ListTags(ctx, services.ListTags{ArticleID: &article.ID})
	if err != nil {
		return nil, err
	}

	favorited, err := app.articlesService.IsFavorite(ctx, user.ID, article.ID)
	if err != nil {
		return nil, err
	}

	favoritesCount, err := app.articlesService.GetFavoritesCount(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	var authorProfile *services.Profile

	if user != nil {
		authorProfile, err = app.profilesService.GetProfile(ctx, *article.AuthorID, &user.ID)
	} else {
		authorProfile, err = app.profilesService.GetProfile(ctx, *article.AuthorID, nil)
	}

	if err != nil {
		return nil, err
	}

	articleResponse := newArticleResponse(article, *articleTags, *favorited, *favoritesCount, *authorProfile)

	return &articleResponse, nil
}
