package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
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

type updateArticleRequest struct {
	Article updateArticleRequestArticle `json:"article"`
}

type updateArticleRequestArticle struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Body        *string `json:"body"`
}

type createCommentRequest struct {
	Comment createCommentRequestComment `json:"comment"`
}

type createCommentRequestComment struct {
	Body string `json:"body"`
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

type multipleArticlesResponse struct {
	Articles      []articleResponseArticle `json:"articles"`
	ArticlesCount int                      `json:"articlesCount"`
}

type ListOfTagsResponse struct {
	Tags []string `json:"tags"`
}

type commentResponse struct {
	Comment commentResponseComment `json:"comment"`
}

type commentResponseComment struct {
	ID        uuid.UUID              `json:"id"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	Body      string                 `json:"body"`
	Author    profileResponseProfile `json:"author"`
}

type multipleCommentsResponse struct {
	Comments []commentResponseComment `json:"comments"`
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

func newMultipleArticlesResponse(articleResponseArticles []articleResponseArticle) multipleArticlesResponse {
	return multipleArticlesResponse{
		Articles:      articleResponseArticles,
		ArticlesCount: len(articleResponseArticles),
	}
}

func newListOfTagsResponse(articleTags []model.ArticleTag) ListOfTagsResponse {
	tagList := make([]string, len(articleTags))
	for i, tag := range articleTags {
		tagList[i] = tag.Name
	}

	return ListOfTagsResponse{
		Tags: tagList,
	}
}

func newCommentResponse(comment model.Comment, authorProfile services.Profile) commentResponse {
	return commentResponse{
		Comment: commentResponseComment{
			ID:        comment.ID,
			CreatedAt: *comment.CreatedAt,
			UpdatedAt: *comment.UpdatedAt,
			Body:      comment.Body,
			Author:    newProfileResponseProfile(authorProfile),
		},
	}
}

func (app *application) listArticles(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	query := r.URL.Query()

	var authorIds *[]uuid.UUID
	authorUsername := query.Get("author")
	if authorUsername != "" {
		authorIds = &[]uuid.UUID{}

		author, err := app.usersService.GetUserByUsername(ctx, authorUsername)
		if err != nil {
			app.writeErrorResponse(ctx, w, err)
			return
		}

		*authorIds = append(*authorIds, author.ID)
	}

	var favoritedByUserId *uuid.UUID
	favoritedByUsername := query.Get("favorited")
	if favoritedByUsername != "" {
		favoritedByUser, err := app.usersService.GetUserByUsername(ctx, favoritedByUsername)
		if err != nil {
			app.writeErrorResponse(ctx, w, err)
			return
		}
		favoritedByUserId = &favoritedByUser.ID
	}

	var tagName *string
	tagParam := query.Get("tag")
	if tagParam != "" {
		tagName = &tagParam
	}

	limit := 20
	limitParam := query.Get("limit")
	if limitParam != "" {
		limitValue, err := strconv.Atoi(limitParam)
		if err != nil {
			app.writeErrorResponse(ctx, w, &malformedRequest{
				msg: fmt.Sprintf("Query parameter 'limit' must be an integer. Received %s", limitParam),
			})
			return
		}
		limit = limitValue
	}

	offset := 0
	offsetParam := query.Get("offset")
	if offsetParam != "" {
		offsetValue, err := strconv.Atoi(offsetParam)
		if err != nil {
			app.writeErrorResponse(ctx, w, &malformedRequest{
				msg: fmt.Sprintf("Query parameter 'offset' must be an integer. Received %s", offsetParam),
			})
			return
		}
		offset = offsetValue
	}

	articles, err := app.articlesService.ListArticles(ctx, services.ListArticles{
		AuthorIDs:         authorIds,
		FavoritedByUserID: favoritedByUserId,
		TagName:           tagName,
		Limit:             &limit,
		Offset:            &offset,
	})
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	multipleArticleResponse, err := app.makeMultipleArticlesResponse(ctx, user, *articles)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, multipleArticleResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) feedArticles(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	query := r.URL.Query()

	limit := 20
	limitParam := query.Get("limit")
	if limitParam != "" {
		limitValue, err := strconv.Atoi(limitParam)
		if err != nil {
			app.writeErrorResponse(ctx, w, &malformedRequest{
				msg: fmt.Sprintf("Query parameter 'limit' must be an integer. Received %s", limitParam),
			})
			return
		}
		limit = limitValue
	}

	offset := 0
	offsetParam := query.Get("offset")
	if offsetParam != "" {
		offsetValue, err := strconv.Atoi(offsetParam)
		if err != nil {
			app.writeErrorResponse(ctx, w, &malformedRequest{
				msg: fmt.Sprintf("Query parameter 'offset' must be an integer. Received %s", offsetParam),
			})
			return
		}
		offset = offsetValue
	}

	var authorIds []uuid.UUID

	followedProfiles, err := app.profilesService.ListFollowedProfiles(ctx, user.ID)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	for _, followedProfile := range *followedProfiles {
		authorIds = append(authorIds, followedProfile.UserID)
	}

	articles, err := app.articlesService.ListArticles(ctx, services.ListArticles{
		AuthorIDs: &authorIds,
		Limit:     &limit,
		Offset:    &offset,
	})
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	multipleArticleResponse, err := app.makeMultipleArticlesResponse(ctx, user, *articles)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, multipleArticleResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) getArticleBySlug(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	articleSlug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, articleSlug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	articleResponse, err := app.makeArticleResponse(ctx, user, *article)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, articleResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) createArticle(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var request createArticleRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	user := app.contextGetUser(r)

	article, err := app.articlesService.CreateArticle(ctx, services.NewCreateArticle(user.ID, request.Article.Title, request.Article.Description, request.Article.Body, request.Article.TagList))
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	articleResponse, err := app.makeArticleResponse(ctx, user, *article)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusCreated, articleResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) updateArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	var request updateArticleRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	user := app.contextGetUser(r)

	articleSlug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, articleSlug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if *article.AuthorID != user.ID {
		app.writeErrorResponse(ctx, w, &forbiddenError{msg: fmt.Sprintf("User %s cannot update article with slug %s", user.Username, article.Slug)})
		return
	}

	article, err = app.articlesService.UpdateArticle(ctx, article.ID, services.UpdateArticle{Title: request.Article.Title, Description: request.Article.Description, Body: request.Article.Body})
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	articleResponse, err := app.makeArticleResponse(ctx, user, *article)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, articleResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) deleteArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	articleSlug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, articleSlug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if *article.AuthorID != user.ID {
		app.writeErrorResponse(ctx, w, &forbiddenError{msg: fmt.Sprintf("User %s cannot delete article with slug %s", user.Username, article.Slug)})
		return
	}

	if err = app.articlesService.DeleteArticle(ctx, article.ID); err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}
}

func (app *application) addCommentToArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	var request createCommentRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	user := app.contextGetUser(r)

	slug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, slug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	comment, err := app.articlesService.CreateComment(ctx, article.ID, user.ID, request.Comment.Body)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	commentResponse, err := app.makeCommentResponse(ctx, *comment, user)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusCreated, commentResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) getCommentsFromArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	slug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, slug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	comments, err := app.articlesService.ListComments(ctx, services.ListComments{ArticleID: &article.ID})
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	multipleCommentsResponse, err := app.makeMultipleCommentsResponse(ctx, *comments, user)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, multipleCommentsResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) deleteComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	slug := ps.ByName("slug")

	_, err := app.articlesService.GetArticleBySlug(ctx, slug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	commentIdString := ps.ByName("commentId")

	commentId, err := uuid.Parse(commentIdString)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	comment, err := app.articlesService.GetCommentById(ctx, commentId)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if *comment.AuthorID != user.ID {
		app.writeErrorResponse(ctx, w, &forbiddenError{msg: fmt.Sprintf("User %s cannot delete comment %s", user.ID, comment.AuthorID)})
		return
	}

	if err = app.articlesService.DeleteComment(ctx, comment.ID); err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}
}

func (app *application) favoriteArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	articleSlug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, articleSlug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = app.articlesService.FavoriteArticle(ctx, user.ID, article.ID); err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	article, err = app.articlesService.GetArticleBySlug(ctx, article.Slug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	articleResponse, err := app.makeArticleResponse(ctx, user, *article)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, articleResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) unfavoriteArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)

	articleSlug := ps.ByName("slug")

	article, err := app.articlesService.GetArticleBySlug(ctx, articleSlug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = app.articlesService.UnfavoriteArticle(ctx, user.ID, article.ID); err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	article, err = app.articlesService.GetArticleBySlug(ctx, article.Slug)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	articleResponse, err := app.makeArticleResponse(ctx, user, *article)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, articleResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) getTags(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	tags, err := app.articlesService.ListTags(ctx, services.ListTags{ArticleID: nil})
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	listOfTagsResponse := newListOfTagsResponse(*tags)

	if err = writeJSON(w, http.StatusOK, listOfTagsResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) makeArticleResponse(ctx context.Context, user *model.Users, article model.Article) (*articleResponse, error) {
	articleTags, err := app.articlesService.ListTags(ctx, services.ListTags{ArticleID: &article.ID})
	if err != nil {
		return nil, err
	}

	favorited := false
	if user != nil {
		isFavorite, err := app.articlesService.IsFavorite(ctx, user.ID, article.ID)
		if err != nil {
			return nil, err
		}
		favorited = *isFavorite
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

	articleResponse := newArticleResponse(article, *articleTags, favorited, *favoritesCount, *authorProfile)

	return &articleResponse, nil
}

func (app *application) makeMultipleArticlesResponse(ctx context.Context, user *model.Users, articles []model.Article) (*multipleArticlesResponse, error) {
	articleResponseArticles := make([]articleResponseArticle, len(articles))

	for i, article := range articles {
		articleResponse, err := app.makeArticleResponse(ctx, user, article)
		if err != nil {
			return nil, err
		}
		articleResponseArticles[i] = articleResponse.Article
	}

	multipleArticleResponse := newMultipleArticlesResponse(articleResponseArticles)

	return &multipleArticleResponse, nil
}

func (app *application) makeCommentResponse(ctx context.Context, comment model.Comment, user *model.Users) (*commentResponse, error) {
	var authorProfile *services.Profile
	var err error

	if user != nil {
		authorProfile, err = app.profilesService.GetProfile(ctx, *comment.AuthorID, &user.ID)
		if err != nil {
			return nil, err
		}
	} else {
		authorProfile, err = app.profilesService.GetProfile(ctx, *comment.AuthorID, nil)
		if err != nil {
			return nil, err
		}
	}

	commentResponse := newCommentResponse(comment, *authorProfile)

	return &commentResponse, nil
}

func (app *application) makeMultipleCommentsResponse(ctx context.Context, comments []model.Comment, user *model.Users) (*multipleCommentsResponse, error) {
	commentResponseComments := make([]commentResponseComment, len(comments))

	for i, comment := range comments {
		commentResponse, err := app.makeCommentResponse(ctx, comment, user)
		if err != nil {
			return nil, err
		}
		commentResponseComments[i] = commentResponse.Comment
	}

	multipleCommentsResponse := multipleCommentsResponse{Comments: commentResponseComments}

	return &multipleCommentsResponse, nil
}
