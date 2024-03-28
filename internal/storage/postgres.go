package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Svirex/microurl/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(ctx context.Context, db *sqlx.DB, migrationsPath string) (*PostgresRepository, error) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("create instance db for migrate: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("create migrate: %w", err)
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("migration up: %w", err)
	}
	return &PostgresRepository{
		db: db,
	}, nil
}

var _ URLRepository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Add(ctx context.Context, d *models.RepositoryAddRecord) (*models.RepositoryGetRecord, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO records (url, short_id) 
										VALUES ($1, $2);`, d.URL, d.ShortID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		row := r.db.QueryRowContext(ctx, "SELECT short_id FROM records WHERE url=$1;", d.URL)
		var shortID string
		err = row.Scan(&shortID)
		if err != nil {
			return nil, fmt.Errorf("select short_id for url: %w", err)
		}
		return models.NewRepositoryGetRecord(shortID), fmt.Errorf("%w", ErrAlreadyExists)
	} else if err != nil {
		return nil, fmt.Errorf("insert row into records: %w", err)
	}
	return models.NewRepositoryGetRecord(d.ShortID), nil
}

func (r *PostgresRepository) Get(ctx context.Context, d *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	row := r.db.QueryRowContext(ctx, "SELECT url FROM records WHERE short_id=$1", d.ShortID)
	var url string
	err := row.Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: postgres get by short id: %w", ErrNotFound, err)
	} else if err != nil {
		return nil, fmt.Errorf("postgres get by short id: %w", err)
	}

	return models.NewRepositoryGetResult(url), nil
}

func (r *PostgresRepository) Shutdown() error {
	return nil
}

func (r *PostgresRepository) Batch(ctx context.Context, batch *models.BatchService) (*models.BatchResponse, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("coulndt begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO records (url, short_id) 
											VALUES ($1, $2) 
											ON CONFLICT (url) DO UPDATE
												SET short_id=records.short_id
									 		RETURNING short_id;`)
	if err != nil {
		return nil, fmt.Errorf("coulndt prepare statement: %w", err)
	}
	response := &models.BatchResponse{
		Records: make([]models.BatchResponseRecord, len(batch.Records)),
	}
	for i := range batch.Records {
		row := stmt.QueryRowContext(ctx, batch.Records[i].URL, batch.Records[i].ShortURL)
		err := row.Scan(&response.Records[i].ShortURL)
		if err != nil {
			return nil, fmt.Errorf("coulndt scan short id: %w", err)
		}
		response.Records[i].CorrID = batch.Records[i].CorrID
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("coulndt commit transaction: %w", err)
	}
	return response, nil
}

func (m *PostgresRepository) UserURLs(_ context.Context, uid string) ([]models.UserURLRecord, error) {
	result := make([]models.UserURLRecord, 0)

	return result, nil
}
