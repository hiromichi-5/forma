package service

import "errors"

var (
	ErrForbidden    = errors.New("forbidden")
	ErrValidation   = errors.New("validation")
	ErrUserNotFound = errors.New("user not found")
)
