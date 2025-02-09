// user-service\internal\domain\user.go

package domain

import "errors"

type User struct {
	ID       string
	Username string
	Email    string
}

var (
	ErrInvalidEmail = errors.New("invalid email")
)
