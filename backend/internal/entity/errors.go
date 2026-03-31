package entity

import "fmt"

type (
	Code      string
	FieldCode string
)

const (
	// 認証・認可
	CodeInvalidCredentials Code = "INVALID_CREDENTIALS" // #nosec
	CodeInvalidSession     Code = "INVALID_SESSION"
	CodeEmailNotVerified   Code = "EMAIL_NOT_VERIFIED"
	CodeForbidden          Code = "FORBIDDEN"
	CodeResourceHidden     Code = "RESOURCE_HIDDEN"

	// リソース不在
	CodeUserNotFound   Code = "USER_NOT_FOUND"
	CodeFormNotFound   Code = "FORM_NOT_FOUND"
	CodeFormNotShared  Code = "FORM_NOT_SHARED"
	CodeTokenNotFound  Code = "TOKEN_NOT_FOUND"
	CodeInviteNotFound Code = "INVITE_NOT_FOUND"

	// ビジネスルール違反
	CodeInviteExpired             Code = "INVITE_EXPIRED"
	CodeAlreadyMember             Code = "ALREADY_MEMBER"
	CodeIncorrectPassword         Code = "INCORRECT_PASSWORD"
	CodeLastAdmin                 Code = "LAST_ADMIN"
	CodeConflict                  Code = "CONFLICT"
	CodeFormAlreadyRegistered     Code = "FORM_ALREADY_REGISTERED"
	CodeActiveInviteAlreadyExists Code = "ACTIVE_INVITE_ALREADY_EXISTS"
	CodeStatusConflict            Code = "STATUS_CONFLICT"

	CodeValidation Code = "VALIDATION_ERROR"
)

const (
	FieldCodeRequired      FieldCode = "REQUIRED"
	FieldCodeTooShort      FieldCode = "TOO_SHORT"
	FieldCodeInvalidFormat FieldCode = "INVALID_FORMAT"
	FieldCodeInvalidValue  FieldCode = "INVALID_VALUE"
)

type Error struct {
	Code   Code
	Fields []FieldError
	Err    error
}

type FieldError struct {
	Field string
	Code  FieldCode
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return string(e.Code)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NewError(code Code) *Error {
	return &Error{Code: code}
}

func NewValidationError(fields ...FieldError) *Error {
	return &Error{Code: CodeValidation, Fields: fields}
}

func WrapError(code Code, err error) *Error {
	return &Error{Code: code, Err: err}
}

func NewFieldError(field string, code FieldCode) FieldError {
	return FieldError{Field: field, Code: code}
}
