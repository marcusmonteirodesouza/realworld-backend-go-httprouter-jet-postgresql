package main

import (
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

	err = writeJSON(w, http.StatusCreated, userResponse, nil)
	if err != nil {
		app.writeErrorResponse(w, err)
	}
}
