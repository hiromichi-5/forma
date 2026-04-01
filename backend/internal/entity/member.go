package entity

import "github.com/google/uuid"

type Member struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	Role        string
}

const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
)
