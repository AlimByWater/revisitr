package auth

import "fmt"

var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrUserExists         = fmt.Errorf("user with this email already exists")
	ErrTokenExpired       = fmt.Errorf("token expired or invalid")
)
