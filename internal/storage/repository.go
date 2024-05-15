package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Svirex/microurl/internal/config"
	"github.com/Svirex/microurl/internal/models"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("not found record")
var ErrSomtheingWrong = errors.New("unknown error")
var ErrAlreadyExists = errors.New("short id already exists")

type URLRepository interface {
	Add(context.Context, *models.RepositoryAddRecord) (*models.RepositoryGetRecord, error)
	Get(context.Context, *models.RepositoryGetRecord) (*models.RepositoryGetResult, error)
	Batch(context.Context, *models.BatchService) (*models.BatchResponse, error)
	UserURLs(ctx context.Context, uid string) ([]models.UserURLRecord, error)
	Shutdown() error
}

func NewRepository(ctx context.Context, cfg *config.Config, db *sqlx.DB) (URLRepository, error) {
	var repository URLRepository
	var err error
	if cfg.PostgresDSN != "" {
		repository, err = NewPostgresRepository(ctx, db, cfg.MigrationsPath)
		if err != nil {
			return nil, fmt.Errorf("create postgres repository: %w", err)
		}
	} else if cfg.FileStoragePath != "" {
		repository, err = NewFileRepository(ctx, cfg.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("create file repository: %w", err)
		}
	} else {
		repository = NewMapRepository()
	}
	return repository, nil
}
