package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	_, err := r.db.ExecContext(ctx, `INSERT INTO records (url, short_id) 
										VALUES ($1, $2);`, d.URL, d.ShortID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		row := r.db.QueryRowContext(ctx, "SELECT short_id FROM records WHERE url=$1;", d.URL)
		var shortID string
		err = row.Scan(&shortID)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return models.NewRepositoryGetRecord(shortID), fmt.Errorf("%w", repositories.ErrAlreadyExists)
	} else if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return models.NewRepositoryGetRecord(d.ShortID), nil
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

func (r *PostgresRepository) Batch(ctx context.Context, batch *models.BatchService) (*models.BatchResponse, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("coulndt add batch. err: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO records (url, short_id) 
											VALUES ($1, $2) 
											ON CONFLICT (url) DO UPDATE
												SET short_id=records.short_id
									 		RETURNING short_id;`)
	if err != nil {
		return nil, fmt.Errorf("coulndt add batch. err: %w", err)
	}
	response := &models.BatchResponse{
		Records: make([]models.BatchResponseRecord, len(batch.Records)),
	}
	for i := range batch.Records {
		row := stmt.QueryRowContext(ctx, batch.Records[i].URL, batch.Records[i].ShortURL)
		err := row.Scan(&response.Records[i].ShortURL)
		if err != nil {
			return nil, fmt.Errorf("coulndt add batch. err: %w", err)
		}
		response.Records[i].CorrID = batch.Records[i].CorrID
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("coulndt add batch. err: %w", err)
	}
	return response, nil
}
