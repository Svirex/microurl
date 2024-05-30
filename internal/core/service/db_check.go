package service

import (
	"context"
	"fmt"

	"github.com/Svirex/microurl/internal/config"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoOpDBCheck struct{}

var _ ports.DBCheck = (*NoOpDBCheck)(nil)

func (n *NoOpDBCheck) Ping(context.Context) error {
	return fmt.Errorf("no op db check, ping")
}

type DBCheck struct {
	db *pgxpool.Pool
}

var _ ports.DBCheck = (*DBCheck)(nil)

func NewDBCheckService(db *pgxpool.Pool) *DBCheck {
	return &DBCheck{
		db: db,
	}
}

func (c *DBCheck) Ping(ctx context.Context) error {
	return c.db.Ping(ctx)
}

func NewDBCheck(db *pgxpool.Pool, cfg *config.Config) ports.DBCheck {
	var dbCheckService ports.DBCheck

	if cfg.PostgresDSN != "" {
		dbCheckService = NewDBCheckService(db)
	} else {
		dbCheckService = &NoOpDBCheck{}
	}
	return dbCheckService
}
