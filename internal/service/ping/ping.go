package ping

import (
	"context"
)

type database interface {
	Ping(ctx context.Context) error
}

type Service struct {
	db database
}

func NewService(db database) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}
