package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/.gen/realworld/public/model"
	"github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/internal/services"
)

type registerUserRequest struct {
	User registerUserRequestUser `json:"user"`
}

type registerUserRequestUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type loginRequest struct {
	User loginRequestUser `json:"user"`
}

type loginRequestUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateUserRequest struct {
	User updateUserRequestUser `json:"user"`
}

type updateUserRequestUser struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
}

type userResponse struct {
	User userResponseUser `json:"user"`
}

type userResponseUser struct {
	Email    string  `json:"email"`
	Token    string  `json:"token"`
	Username string  `json:"username"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
}

func newUserResponse(user model.Users, token string) userResponse {
	return userResponse{
		User: userResponseUser{
			Email:    user.Email,
			Token:    token,
			Username: user.Username,
			Bio:      user.Bio,
			Image:    user.Image,
		},
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var request loginRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	user, err := app.usersService.GetUserByEmail(ctx, request.User.Email)
	if err != nil {
		if errNotFound, ok := err.(*services.NotFoundError); ok {
			app.writeErrorResponse(ctx, w, &unauthorizedError{msg: errNotFound.Error()})
		} else {
			app.writeErrorResponse(ctx, w, err)
		}
		return
	}

	isCorrectPassword, err := app.usersService.CheckPassword(ctx, user.ID, request.User.Password)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	if !(*isCorrectPassword) {
		app.writeErrorResponse(ctx, w, &unauthorizedError{msg: fmt.Sprintf("Incorrect password for user %s", user.ID)})
		return
	}

	token, err := app.usersService.GetToken(ctx, user)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	userResponse := newUserResponse(*user, *token)

	if err = writeJSON(w, http.StatusOK, userResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) registerUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var request registerUserRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	user, err := app.usersService.RegisterUser(ctx, services.NewRegisterUser(request.User.Email, request.User.Username, request.User.Password))
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	token, err := app.usersService.GetToken(ctx, user)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	userResponse := newUserResponse(*user, *token)

	if err = writeJSON(w, http.StatusCreated, userResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) getCurrentUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	user := app.contextGetUser(r)
	token := app.contextGetToken(r)

	userResponse := newUserResponse(*user, token)

	err := writeJSON(w, http.StatusOK, userResponse)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	var request updateUserRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	user := app.contextGetUser(r)

	token := app.contextGetToken(r)

	user, err = app.usersService.UpdateUser(ctx, user.ID, services.UpdateUser{Email: request.User.Email, Username: request.User.Username, Password: request.User.Password, Bio: request.User.Bio, Image: request.User.Image})
	if err != nil {
		app.writeErrorResponse(ctx, w, err)
		return
	}

	userResponse := newUserResponse(*user, token)

	if err = writeJSON(w, http.StatusOK, userResponse); err != nil {
		app.writeErrorResponse(ctx, w, err)
	}
}
