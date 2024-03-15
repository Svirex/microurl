package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	url TEXT UNIQUE NOT NULL,
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

func (r *PostgresRepository) Add(ctx context.Context, d *models.RepositoryAddRecord) (*models.RepositoryGetRecord, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO records (url, short_id) 
									VALUES ($1, $2) 
									ON CONFLICT (url) DO UPDATE
									SET short_id=records.short_id
									 RETURNING short_id;
	`, d.URL, d.ShortID)
	var short_id string
	err := row.Scan(&short_id)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("couldnt insert record")
	}
	return models.NewRepositoryGetRecord(short_id), nil
}

func (r *PostgresRepository) Get(ctx context.Context, d *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	row := r.db.QueryRowContext(ctx, "SELECT url FROM records WHERE short_id=$1", d.ShortID)
	var url string
	err := row.Scan(&url)
	if err == sql.ErrNoRows {
		return nil, repositories.ErrNotFound
	} else if err != nil {
		return nil, repositories.ErrSomtheingWrong
	}

	return models.NewRepositoryGetResult(url), nil
}

func (r *PostgresRepository) Shutdown() error {
	return nil
}
