package service

import (
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/hiromichi-5/forma/backend/internal/google"
)

type Service struct {
	Q  *db.Queries
	GF google.FormsClient
}

func NewService(q *db.Queries, gf google.FormsClient) *Service {
	return &Service{Q: q, GF: gf}
}
