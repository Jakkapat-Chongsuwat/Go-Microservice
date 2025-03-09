package domain

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

var (
	ErrInvalidEmail = errors.New("invalid email")
)

func NewUser(username, email string) *User {
	return &User{
		ID:       uuid.NewString(),
		Username: username,
		Email:    email,
	}
}
