package services

import (
	"context"
	"database/sql"
	// . "github.com/marcusmonteirodesouza/realworld-backend-go-jet-postgresql/internal/users/models"
)

type UsersService struct {
	db *sql.DB
}

func NewUsersService(db *sql.DB) *UsersService {
	return &UsersService{
		db: db,
	}
}

func (usersService *UsersService) RegisterUser(ctx context.Context, registerUser RegisterUser) error {
	return nil
}
