package services

import (
	"context"
	"database/sql"
	"log/slog"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/model"
	. "github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/table"
)

type ProfilesService struct {
	db           *sql.DB
	logger       *slog.Logger
	usersService *UsersService
}

func NewProfilesService(db *sql.DB, logger *slog.Logger, usersService *UsersService) ProfilesService {
	return ProfilesService{
		db:           db,
		logger:       logger,
		usersService: usersService,
	}
}

type Profile struct {
	UserID    uuid.UUID
	Username  string
	Bio       *string
	Image     *string
	Following bool
}

func NewProfile(user model.Users, following bool) Profile {
	return Profile{
		UserID:    user.ID,
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
		Following: following,
	}
}

func (profilesService *ProfilesService) GetProfile(ctx context.Context, userId uuid.UUID, followerId *uuid.UUID) (*Profile, error) {
	user, err := profilesService.usersService.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}

	following := false
	if followerId != nil {
		isFollowing, err := profilesService.IsFollowing(ctx, *followerId, user.ID)
		if err != nil {
			return nil, err
		}
		following = *isFollowing
	}

	profile := NewProfile(*user, following)

	return &profile, nil
}

func (profilesService *ProfilesService) FollowUser(ctx context.Context, followerId uuid.UUID, followedId uuid.UUID) error {
	profilesService.logger.InfoContext(ctx, "Following user", "followerId", followerId, "followedId", followedId)

	isFollowing, err := profilesService.IsFollowing(ctx, followerId, followedId)
	if err != nil {
		return err
	}

	if *isFollowing {
		return nil
	}

	follower, err := profilesService.usersService.GetUserById(ctx, followerId)
	if err != nil {
		return err
	}

	followed, err := profilesService.usersService.GetUserById(ctx, followedId)
	if err != nil {
		return err
	}

	follow := model.Follow{
		FollowerID: &follower.ID,
		FollowedID: &followed.ID,
	}

	followUserStmt := Follow.INSERT(Follow.FollowerID, Follow.FollowedID).MODEL(follow).RETURNING(Follow.AllColumns)

	if err = followUserStmt.QueryContext(ctx, profilesService.db, &follow); err != nil {
		return err
	}

	return nil
}

func (profilesService *ProfilesService) UnfollowUser(ctx context.Context, followerId uuid.UUID, followedId uuid.UUID) error {
	profilesService.logger.InfoContext(ctx, "Unfollowing user", "followerId", followerId, "followedId", followedId)

	isFollowing, err := profilesService.IsFollowing(ctx, followerId, followedId)
	if err != nil {
		return err
	}

	if !*isFollowing {
		return nil
	}

	unfollowUserStmt := Follow.DELETE().WHERE(Follow.FollowerID.EQ(UUID(followerId)).AND(Follow.FollowedID.EQ(UUID(followedId))))

	_, err = unfollowUserStmt.ExecContext(ctx, profilesService.db)
	if err != nil {
		return err
	}

	return nil
}

func (profilesService *ProfilesService) ListFollowedProfiles(ctx context.Context, userId uuid.UUID) (*[]Profile, error) {
	var followedIds []uuid.UUID

	var follows []model.Follow

	listFollowedIdsStmt := SELECT(Follow.FollowedID).FROM(Follow).WHERE(Follow.FollowerID.EQ(UUID(userId)))

	err := listFollowedIdsStmt.QueryContext(ctx, profilesService.db, &follows)
	if err != nil {
		return nil, err
	}

	for _, follow := range follows {
		followedIds = append(followedIds, *follow.FollowedID)
	}

	users, err := profilesService.usersService.ListUsers(ctx, ListUsers{UserIDs: &followedIds})
	if err != nil {
		return nil, err
	}

	profiles := []Profile{}

	for _, user := range *users {
		profiles = append(profiles, NewProfile(user, true))
	}

	return &profiles, nil
}

func (profilesService *ProfilesService) IsFollowing(ctx context.Context, followerId uuid.UUID, followedId uuid.UUID) (*bool, error) {
	var dest struct {
		IsFollowing bool
	}

	isFollowingStmt := SELECT(EXISTS(Follow.SELECT(Follow.ID).WHERE(Follow.FollowerID.EQ(UUID(followerId)).AND(Follow.FollowedID.EQ(UUID(followedId))))).AS("is_following"))

	err := isFollowingStmt.QueryContext(ctx, profilesService.db, &dest)
	if err != nil {
		return nil, err
	}

	return &dest.IsFollowing, nil
}
