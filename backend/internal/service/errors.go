package service

import "errors"

var (
	ErrForbidden         = errors.New("forbidden")
	ErrValidation        = errors.New("validation")
	ErrUserNotFound      = errors.New("user not found")
	ErrConflict          = errors.New("conflict")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrEmailNotVerified  = errors.New("email not verified")
	ErrTokenNotFound     = errors.New("token not found")
	ErrInviteNotFound    = errors.New("invite not found")
	ErrInviteExpired     = errors.New("invite expired")
	ErrInviteRevoked     = errors.New("invite revoked")
	ErrAlreadyMember     = errors.New("already member")
	ErrCodeGeneration    = errors.New("code generation failed")
)
