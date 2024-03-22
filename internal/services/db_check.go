package services

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrPingFailed = errors.New("ping to db failed")

type DBCheck interface {
	Ping(context.Context) error
	Shutdown() error
}

type NoOpDBCheck struct{}

var _ DBCheck = (*NoOpDBCheck)(nil)

func (n *NoOpDBCheck) Ping(context.Context) error {
	return ErrPingFailed
}

func (n *NoOpDBCheck) Shutdown() error {
	return nil
}

type DefaultDBCheck struct {
	db *sqlx.DB
}

var _ DBCheck = (*DefaultDBCheck)(nil)

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
