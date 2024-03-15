package services

import (
	"context"

	"github.com/Svirex/microurl/internal/pkg/services"
	"github.com/jmoiron/sqlx"
)

type DefaultDBCheck struct {
	db *sqlx.DB
}

var _ services.DBCheck = (*DefaultDBCheck)(nil)

func NewDBCheckService(db *sqlx.DB) *DefaultDBCheck {
	return &DefaultDBCheck{
		db: db,
	}
}

func (c *DefaultDBCheck) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func (c *DefaultDBCheck) Shutdown() error {
	return c.db.Close()
}
