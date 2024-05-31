package service

import (
	"context"

	"github.com/Svirex/microurl/internal/config"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NoOpDBCheck - заглушка для сервиса проверки соединения к БД.
type NoOpDBCheck struct{}

var _ ports.DBCheck = (*NoOpDBCheck)(nil)

// Ping - проверяем соединение.
func (n *NoOpDBCheck) Ping(context.Context) error {
	return nil
}

// DBCheck - структура сервиса.
type DBCheck struct {
	db *pgxpool.Pool
}

var _ ports.DBCheck = (*DBCheck)(nil)

// NewDBCheckService - новый сервис.
func NewDBCheckService(db *pgxpool.Pool) *DBCheck {
	return &DBCheck{
		db: db,
	}
}

// Ping - проверяем соединение.
func (c *DBCheck) Ping(ctx context.Context) error {
	return c.db.Ping(ctx)
}

// NewDBCheck - новый сервис на основе конфига.
func NewDBCheck(db *pgxpool.Pool, cfg *config.Config) ports.DBCheck {
	var dbCheckService ports.DBCheck

	if cfg.PostgresDSN != "" {
		dbCheckService = NewDBCheckService(db)
	} else {
		dbCheckService = &NoOpDBCheck{}
	}
	return dbCheckService
}
