package models

import "time"

type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	Bio          *string
	Image        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
