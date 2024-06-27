package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/internal/services"
)

type profileResponse struct {
	Profile profileResponseProfile `json:"profile"`
}

type profileResponseProfile struct {
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

func newProfileResponse(profile services.Profile) profileResponse {
	return profileResponse{
		Profile: newProfileResponseProfile(profile),
	}
}

func newProfileResponseProfile(profile services.Profile) profileResponseProfile {
	return profileResponseProfile{
		Username:  profile.Username,
		Bio:       profile.Bio,
		Image:     profile.Image,
		Following: profile.Following,
	}
}

func (app *application) getProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	username := ps.ByName("username")

	user, err := app.usersService.GetUserByUsername(ctx, username)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	var followerId uuid.UUID
	follower := app.contextGetUser(r)
	if follower != nil {
		followerId = follower.ID
	}

	profile, err := app.profilesService.GetProfile(ctx, user.ID, &followerId)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, newProfileResponse(*profile)); err != nil {
		app.writeErrorResponse(w, err)
	}
}

func (app *application) followUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	username := ps.ByName("username")

	user, err := app.usersService.GetUserByUsername(ctx, username)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	follower := app.contextGetUser(r)

	if err = app.profilesService.FollowUser(ctx, follower.ID, user.ID); err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	profile, err := app.profilesService.GetProfile(ctx, user.ID, &follower.ID)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, newProfileResponse(*profile)); err != nil {
		app.writeErrorResponse(w, err)
	}
}

func (app *application) unfollowUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	username := ps.ByName("username")

	user, err := app.usersService.GetUserByUsername(ctx, username)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	follower := app.contextGetUser(r)

	if err = app.profilesService.UnfollowUser(ctx, follower.ID, user.ID); err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	profile, err := app.profilesService.GetProfile(ctx, user.ID, &follower.ID)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	if err = writeJSON(w, http.StatusOK, newProfileResponse(*profile)); err != nil {
		app.writeErrorResponse(w, err)
	}
}
