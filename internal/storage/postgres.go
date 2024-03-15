package storage

import (
	"context"

	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

var schema = `
CREATE TABLE IF NOT EXISTS
public.records (
	id SERIAL PRIMARY KEY,
	url TEXT NOT NULL UNIQUE,
	short_id VARCHAR(32) NOT NULL
)
`

func NewPostgresRepository(ctx context.Context, db *sqlx.DB) (*PostgresRepository, error) {
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{
		db: db,
	}, nil
}

var _ repositories.URLRepository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Add(context.Context, *models.RepositoryAddRecord) error {
	return nil
}

func (r *PostgresRepository) Get(context.Context, *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	return models.NewRepositoryGetResult("FAKE"), nil
}

func (r *PostgresRepository) Shutdown() error {
	return nil
}
