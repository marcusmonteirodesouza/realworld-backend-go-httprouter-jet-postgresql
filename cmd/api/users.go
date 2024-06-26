package main

import (
	"fmt"
	"net/http"

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

func newUserResponse(email string, token string, username string, bio *string, image *string) userResponse {
	return userResponse{
		User: userResponseUser{
			Email:    email,
			Token:    token,
			Username: username,
			Bio:      bio,
			Image:    image,
		},
	}
}

func (app *application) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	token := app.contextGetToken(r)

	userResponse := newUserResponse(user.Email, token, user.Username, user.Bio, user.Image)

	err := writeJSON(w, http.StatusOK, userResponse)
	if err != nil {
		app.writeErrorResponse(w, err)
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	ctx := r.Context()

	user, err := app.usersService.GetUserByEmail(ctx, request.User.Email)
	if err != nil {
		if errNotFound, ok := err.(*services.NotFoundError); ok {
			app.writeErrorResponse(w, &unauthorizedError{msg: errNotFound.Error()})
		} else {
			app.writeErrorResponse(w, err)
		}
		return
	}

	isCorrectPassword, err := app.usersService.CheckPassword(ctx, user.ID, request.User.Password)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	if !(*isCorrectPassword) {
		app.writeErrorResponse(w, &unauthorizedError{msg: fmt.Sprintf("Incorrect password for user %s", user.ID)})
		return
	}

	token, err := app.usersService.GetToken(user)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	userResponse := newUserResponse(user.Email, *token, user.Username, user.Bio, user.Image)

	err = writeJSON(w, http.StatusOK, userResponse)
	if err != nil {
		app.writeErrorResponse(w, err)
	}
}

func (app *application) registerUser(w http.ResponseWriter, r *http.Request) {
	var request registerUserRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	ctx := r.Context()

	user, err := app.usersService.RegisterUser(ctx, services.NewRegisterUser(request.User.Email, request.User.Username, request.User.Password))
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	token, err := app.usersService.GetToken(user)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	userResponse := newUserResponse(user.Email, *token, user.Username, user.Bio, user.Image)

	err = writeJSON(w, http.StatusCreated, userResponse)
	if err != nil {
		app.writeErrorResponse(w, err)
	}
}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {
	var request updateUserRequest

	err := decodeJSONBody(w, r, &request)
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	ctx := r.Context()

	user := app.contextGetUser(r)

	token := app.contextGetToken(r)

	updatedUser, err := app.usersService.UpdateUser(ctx, user.ID, services.NewUpdateUser(request.User.Email, request.User.Username, request.User.Password, request.User.Bio, request.User.Image))
	if err != nil {
		app.writeErrorResponse(w, err)
		return
	}

	userResponse := newUserResponse(updatedUser.Email, token, updatedUser.Username, updatedUser.Bio, updatedUser.Image)

	err = writeJSON(w, http.StatusCreated, userResponse)
	if err != nil {
		app.writeErrorResponse(w, err)
	}
}
