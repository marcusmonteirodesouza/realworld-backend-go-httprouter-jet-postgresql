package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/model"
	. "github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/table"
)

type ArticlesService struct {
	db           *sql.DB
	logger       *slog.Logger
	usersService *UsersService
}

func NewArticlesService(db *sql.DB, logger *slog.Logger, usersService *UsersService) ArticlesService {
	return ArticlesService{
		db:           db,
		logger:       logger,
		usersService: usersService,
	}
}

type CreateArticle struct {
	AuthorID    uuid.UUID
	Title       string
	Description string
	Body        string
	TagList     *[]string
}

func NewCreateArticle(authorId uuid.UUID, title string, description string, body string, tagList *[]string) CreateArticle {
	return CreateArticle{
		AuthorID:    authorId,
		Title:       title,
		Description: description,
		Body:        body,
		TagList:     tagList,
	}
}

type ListArticles struct {
	AuthorIDs         *[]uuid.UUID
	FavoritedByUserID *uuid.UUID
	TagName           *string
	Limit             *int
	Offset            *int
}

type UpdateArticle struct {
	Title       *string
	Description *string
	Body        *string
}

type ListTags struct {
	ArticleID *uuid.UUID
}

type ListComments struct {
	ArticleID *uuid.UUID
}

func (articlesService *ArticlesService) CreateArticle(ctx context.Context, createArticle CreateArticle) (*model.Article, error) {
	articlesService.logger.InfoContext(ctx, "Creating article", "authorID", createArticle.AuthorID, "title", createArticle.Title, "description", createArticle.Description, "body", createArticle.Body, "tagList", createArticle.TagList)

	author, err := articlesService.usersService.GetUserById(ctx, createArticle.AuthorID)
	if err != nil {
		return nil, err
	}

	slug, err := articlesService.makeSlug(ctx, author.Username, createArticle.Title)
	if err != nil {
		return nil, err
	}

	article := model.Article{
		AuthorID:    &author.ID,
		Slug:        *slug,
		Title:       createArticle.Title,
		Description: createArticle.Description,
		Body:        createArticle.Body,
	}

	tx, err := articlesService.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	articleInsertStmt := Article.INSERT(Article.AuthorID, Article.Slug, Article.Title, Article.Description, Article.Body).MODEL(article).RETURNING(Article.AllColumns)

	if err = articleInsertStmt.QueryContext(ctx, tx, &article); err != nil {
		return nil, err
	}

	if createArticle.TagList != nil {
		for _, tagName := range *createArticle.TagList {
			tagName = articlesService.makeTagName(tagName)

			tag, err := articlesService.getTagByName(ctx, tagName)
			if err != nil {
				if _, ok := err.(*NotFoundError); ok {
					tagModel := model.ArticleTag{
						Name: tagName,
					}

					insertTagStmt := ArticleTag.INSERT(ArticleTag.Name).MODEL(tagModel).RETURNING(ArticleTag.AllColumns)

					articlesService.logger.InfoContext(ctx, "Creating article tag", "name", tagModel.Name)

					if err = insertTagStmt.QueryContext(ctx, tx, &tagModel); err != nil {
						return nil, err
					}

					tag = &tagModel
				} else {
					return nil, err
				}
			}

			articleArticleTag := model.ArticleArticleTag{
				ArticleID:    &article.ID,
				ArticleTagID: &tag.ID,
			}

			insertArticleArticleTagStmt := ArticleArticleTag.INSERT(ArticleArticleTag.ArticleID, ArticleArticleTag.ArticleTagID).MODEL(articleArticleTag)

			if _, err = insertArticleArticleTagStmt.ExecContext(ctx, tx); err != nil {
				return nil, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	articlesService.logger.InfoContext(ctx, "Article created", "articleId", article.ID, "slug", article.Slug)

	return &article, nil
}

func (articlesService *ArticlesService) GetArticleById(ctx context.Context, articleId uuid.UUID) (*model.Article, error) {
	var article model.Article

	getArticleByIdStmt := Article.SELECT(Article.AllColumns).WHERE(Article.ID.EQ(UUID(articleId)))

	err := getArticleByIdStmt.QueryContext(ctx, articlesService.db, &article)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, &NotFoundError{msg: fmt.Sprintf("Article %s not found", articleId)}
		}
		return nil, err
	}

	return &article, nil
}

func (articlesService *ArticlesService) GetArticleBySlug(ctx context.Context, slug string) (*model.Article, error) {
	var article model.Article

	getArticleBySlugStmt := Article.SELECT(Article.AllColumns).WHERE(Article.Slug.EQ(String(slug)))

	err := getArticleBySlugStmt.QueryContext(ctx, articlesService.db, &article)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, &NotFoundError{msg: fmt.Sprintf("Article with slug %s not found", slug)}
		}
		return nil, err
	}

	return &article, nil
}

func (articlesService *ArticlesService) ListArticles(ctx context.Context, listArticles ListArticles) (*[]model.Article, error) {
	condition := Bool(true)

	if listArticles.AuthorIDs != nil {
		if len(*listArticles.AuthorIDs) > 0 {
			var sqlAuthorIds []Expression

			for _, authorId := range *listArticles.AuthorIDs {
				sqlAuthorIds = append(sqlAuthorIds, UUID(authorId))
			}

			condition = condition.AND(Article.AuthorID.IN(sqlAuthorIds...))
		} else {
			condition = condition.AND(Bool(false))
		}
	}

	if listArticles.FavoritedByUserID != nil {
		condition = condition.AND(Article.ID.IN(SELECT(ArticleFavorite.ArticleID).FROM(ArticleFavorite).WHERE(ArticleFavorite.UserID.EQ(UUID(*listArticles.FavoritedByUserID)))))
	}

	if listArticles.TagName != nil {
		condition = condition.AND(
			Article.ID.IN(
				SELECT(ArticleArticleTag.ArticleID).FROM(
					ArticleArticleTag.INNER_JOIN(ArticleTag, ArticleArticleTag.ArticleTagID.EQ(ArticleTag.ID))).WHERE(ArticleTag.Name.EQ(String(*listArticles.TagName)))))
	}

	listArticlesStmt := SELECT(Article.AllColumns).FROM(Article).WHERE(condition).ORDER_BY(Article.CreatedAt.DESC())

	if listArticles.Limit != nil {
		listArticlesStmt = listArticlesStmt.LIMIT(int64(*listArticles.Limit))
	}

	if listArticles.Offset != nil {
		listArticlesStmt = listArticlesStmt.OFFSET(int64(*listArticles.Offset))
	}

	var articles []model.Article

	err := listArticlesStmt.QueryContext(ctx, articlesService.db, &articles)
	if err != nil {
		return nil, err
	}

	return &articles, nil
}

func (articlesService *ArticlesService) UpdateArticle(ctx context.Context, articleId uuid.UUID, updateArticle UpdateArticle) (*model.Article, error) {
	articlesService.logger.InfoContext(ctx, "Updating article", "articleId", articleId, "title", updateArticle.Title, "description", updateArticle.Description, "body", updateArticle.Body)

	article, err := articlesService.GetArticleById(ctx, articleId)
	if err != nil {
		return nil, err
	}

	if updateArticle.Title != nil {
		author, err := articlesService.usersService.GetUserById(ctx, *article.AuthorID)
		if err != nil {
			return nil, err
		}

		slug, err := articlesService.makeSlug(ctx, author.Username, *updateArticle.Title)
		if err != nil {
			return nil, err
		}

		article.Slug = *slug
		article.Title = *updateArticle.Title
	}

	if updateArticle.Description != nil {
		article.Description = *updateArticle.Description
	}

	if updateArticle.Body != nil {
		article.Body = *updateArticle.Body
	}

	now := time.Now().UTC()

	article.UpdatedAt = &now

	updateArticleStmt := Article.UPDATE(Article.Slug, Article.Title, Article.Description, Article.Body, Article.UpdatedAt).MODEL(article).WHERE(Article.ID.EQ(UUID(article.ID)))

	_, err = updateArticleStmt.ExecContext(ctx, articlesService.db)
	if err != nil {
		return nil, err
	}

	return article, nil
}

func (articlesService *ArticlesService) DeleteArticle(ctx context.Context, articleId uuid.UUID) error {
	articlesService.logger.InfoContext(ctx, "Deleting article", "articleId", articleId)

	deleteArticleStmt := Article.DELETE().WHERE(Article.ID.EQ(UUID(articleId)))

	sqlResult, err := deleteArticleStmt.ExecContext(ctx, articlesService.db)
	if err != nil {
		return err
	}

	rowsAffected, err := sqlResult.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return &NotFoundError{msg: fmt.Sprintf("Article %s not found", articleId.String())}
	}

	return nil
}

func (articlesService *ArticlesService) CreateComment(ctx context.Context, articleId uuid.UUID, authorId uuid.UUID, body string) (*model.ArticleComment, error) {
	articlesService.logger.InfoContext(ctx, "Creating comment", "articleId", articleId, "authorId", authorId, "body", body)

	article, err := articlesService.GetArticleById(ctx, articleId)
	if err != nil {
		return nil, err
	}

	author, err := articlesService.usersService.GetUserById(ctx, authorId)
	if err != nil {
		return nil, err
	}

	comment := model.ArticleComment{
		ArticleID: &article.ID,
		AuthorID:  &author.ID,
		Body:      body,
	}

	addCommentStmt := ArticleComment.INSERT(ArticleComment.ArticleID, ArticleComment.AuthorID, ArticleComment.Body).MODEL(comment).RETURNING(ArticleComment.AllColumns)
	if err = addCommentStmt.QueryContext(ctx, articlesService.db, &comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

func (articlesService *ArticlesService) GetCommentById(ctx context.Context, commentId uuid.UUID) (*model.ArticleComment, error) {
	var comment model.ArticleComment

	getCommentStmt := SELECT(ArticleComment.AllColumns).FROM(ArticleComment).WHERE(ArticleComment.ID.EQ(UUID(commentId)))

	err := getCommentStmt.QueryContext(ctx, articlesService.db, &comment)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, &NotFoundError{msg: fmt.Sprintf("Comment %s not found", commentId)}
		}
		return nil, err
	}

	return &comment, nil
}

func (articlesService *ArticlesService) ListComments(ctx context.Context, listComments ListComments) (*[]model.ArticleComment, error) {
	condition := Bool(true)

	if listComments.ArticleID != nil {
		condition = condition.AND(ArticleComment.ArticleID.EQ(UUID(listComments.ArticleID)))
	}

	var comments []model.ArticleComment

	listCommentsStmt := SELECT(ArticleComment.AllColumns).FROM(ArticleComment).WHERE(condition).ORDER_BY(ArticleComment.CreatedAt.DESC())

	err := listCommentsStmt.QueryContext(ctx, articlesService.db, &comments)
	if err != nil {
		return nil, err
	}

	return &comments, nil
}

func (articlesService *ArticlesService) DeleteComment(ctx context.Context, commentId uuid.UUID) error {
	articlesService.logger.InfoContext(ctx, "Deleting comment", "commentId", commentId)

	deleteCommentStmt := ArticleComment.DELETE().WHERE(ArticleComment.ID.EQ(UUID(commentId)))

	deleteCommentSqlResult, err := deleteCommentStmt.ExecContext(ctx, articlesService.db)
	if err != nil {
		return err
	}

	rowsAffected, err := deleteCommentSqlResult.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return &NotFoundError{msg: fmt.Sprintf("Comment %s not found", commentId)}
	}

	return nil
}

func (articlesService *ArticlesService) FavoriteArticle(ctx context.Context, userId uuid.UUID, articleId uuid.UUID) error {
	articlesService.logger.InfoContext(ctx, "Favoriting article", "userId", userId, "articleId", articleId)

	isFavorite, err := articlesService.IsFavorite(ctx, userId, articleId)
	if err != nil {
		return err
	}

	if *isFavorite {
		return nil
	}

	user, err := articlesService.usersService.GetUserById(ctx, userId)
	if err != nil {
		return err
	}

	article, err := articlesService.GetArticleById(ctx, articleId)
	if err != nil {
		return err
	}

	favorite := model.ArticleFavorite{
		UserID:    &user.ID,
		ArticleID: &article.ID,
	}

	insertFavoriteStmt := ArticleFavorite.INSERT(ArticleFavorite.UserID, ArticleFavorite.ArticleID).MODEL(favorite)

	_, err = insertFavoriteStmt.ExecContext(ctx, articlesService.db)
	if err != nil {
		return err
	}

	return nil
}

func (articlesService *ArticlesService) UnfavoriteArticle(ctx context.Context, userId uuid.UUID, articleId uuid.UUID) error {
	articlesService.logger.InfoContext(ctx, "Unfavoriting article", "userId", userId, "articleId", articleId)

	isFavorite, err := articlesService.IsFavorite(ctx, userId, articleId)
	if err != nil {
		return err
	}

	if !*isFavorite {
		return nil
	}

	deleteFavoriteStmt := ArticleFavorite.DELETE().WHERE(ArticleFavorite.UserID.EQ(UUID(userId)).AND(ArticleFavorite.ArticleID.EQ(UUID(articleId))))

	_, err = deleteFavoriteStmt.ExecContext(ctx, articlesService.db)
	if err != nil {
		return err
	}

	return nil
}

func (articlesService *ArticlesService) IsFavorite(ctx context.Context, userId uuid.UUID, articleId uuid.UUID) (*bool, error) {
	var isFavoriteDest struct {
		IsFavorite bool
	}

	isFavoriteStmt := SELECT(EXISTS(ArticleFavorite.SELECT(ArticleFavorite.ID).WHERE(ArticleFavorite.UserID.EQ(UUID(userId)).AND(ArticleFavorite.ArticleID.EQ(UUID(articleId))))).AS("is_favorite"))

	err := isFavoriteStmt.QueryContext(ctx, articlesService.db, &isFavoriteDest)
	if err != nil {
		return nil, err
	}

	return &isFavoriteDest.IsFavorite, nil
}

func (articlesService *ArticlesService) GetFavoritesCount(ctx context.Context, articleId uuid.UUID) (*int, error) {
	var favoritesCountDest struct {
		FavoritesCount int
	}

	isFavoriteStmt := SELECT(COUNT(STAR).AS("favorites_count")).FROM(ArticleFavorite).WHERE(ArticleFavorite.ArticleID.EQ(UUID(articleId)))

	err := isFavoriteStmt.QueryContext(ctx, articlesService.db, &favoritesCountDest)
	if err != nil {
		return nil, err
	}

	return &favoritesCountDest.FavoritesCount, nil
}

func (articlesService *ArticlesService) ListTags(ctx context.Context, listTags ListTags) (*[]model.ArticleTag, error) {
	var tags []model.ArticleTag

	condition := Bool(true)

	if listTags.ArticleID != nil {
		condition = condition.AND(ArticleTag.ID.IN(ArticleArticleTag.SELECT(ArticleArticleTag.ArticleTagID).WHERE(ArticleArticleTag.ArticleID.EQ(UUID(listTags.ArticleID)))))
	}

	listTagsStmt := SELECT(ArticleTag.AllColumns).FROM(ArticleTag).WHERE(condition).ORDER_BY(ArticleTag.Name)

	err := listTagsStmt.QueryContext(ctx, articlesService.db, &tags)
	if err != nil {
		return nil, err
	}

	return &tags, nil
}

func (articlesService *ArticlesService) getTagByName(ctx context.Context, tagName string) (*model.ArticleTag, error) {
	var tag model.ArticleTag

	getTagByNameStmt := ArticleTag.SELECT(ArticleTag.AllColumns).WHERE(ArticleTag.Name.EQ(String(tagName)))

	err := getTagByNameStmt.QueryContext(ctx, articlesService.db, &tag)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, &NotFoundError{msg: fmt.Sprintf("Tag name %s not found", tagName)}
		}
		return nil, err
	}

	return &tag, nil
}

func (articlesService *ArticlesService) makeSlug(ctx context.Context, authorUsername string, title string) (*string, error) {
	slug := slug.Make(fmt.Sprintf("%s %s", authorUsername, title))

	var slugExistsDest struct {
		SlugExists bool
	}

	slugExistsStmt := SELECT(EXISTS(Article.SELECT(Article.ID).WHERE(Article.Slug.EQ(String(slug)))).AS("slug_exists"))

	err := slugExistsStmt.QueryContext(ctx, articlesService.db, &slugExistsDest)
	if err != nil {
		return nil, err
	}

	if slugExistsDest.SlugExists {
		return nil, &AlreadyExistsError{msg: fmt.Sprintf("Slug %s already exists. Please choose another title.", slug)}
	}

	return &slug, nil
}

func (articlesService *ArticlesService) makeTagName(tagName string) string {
	return slug.Make(tagName)
}
