package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db     *pgxpool.Pool
	logger ports.Logger
}

func NewPostgresRepository(db *pgxpool.Pool, logger ports.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

var _ ports.ShortenerRepository = (*PostgresRepository)(nil)

func (repo *PostgresRepository) Add(ctx context.Context, shortID domain.ShortID, data *domain.Record) (domain.ShortID, error) {
	trx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return shortID, fmt.Errorf("postgres repository, add, start transaction: %w", err)
	}
	defer trx.Rollback(ctx)
	var id int
	err = trx.QueryRow(ctx, `INSERT INTO records (url, short_id) 
							 VALUES ($1, $2) RETURNING id;`, data.URL, shortID).Scan(&id)
	if err != nil {
		trx.Rollback(ctx)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var shortID domain.ShortID
			err = repo.db.QueryRow(ctx, "SELECT short_id FROM records WHERE url=$1;", data.URL).Scan(&shortID)
			if err != nil {
				return shortID, fmt.Errorf("postgres repository, add, select short id: %w", err)
			}
			return shortID, ports.ErrAlreadyExists
		}
		return shortID, fmt.Errorf("postgres repository, add, insert url and short id: %w", err)
	}
	_, err = trx.Exec(ctx, `INSERT INTO users (uid, record_id)
							VALUES ($1, $2)`, data.UID, id)
	if err != nil {
		return shortID, fmt.Errorf("postgres repository, add, insert into users: %w", err)
	}
	err = trx.Commit(ctx)
	if err != nil {
		return shortID, fmt.Errorf("postgres repository, add, commit trx: %w", err)
	}
	return shortID, nil
}

func (repo *PostgresRepository) Get(ctx context.Context, shortID domain.ShortID) (domain.URL, error) {
	var url domain.URL
	err := repo.db.QueryRow(ctx, "SELECT url FROM records WHERE short_id=$1 AND is_deleted=false;", shortID).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repo.logger.Errorf("postgres repository, get, not found: %v", err)
			return url, ports.ErrNotFound
		}
		return url, fmt.Errorf("postgres repository, get, select url: %w", err)
	}
	return url, nil
}

func (repo *PostgresRepository) Batch(ctx context.Context, uid domain.UID, data []domain.BatchRecord) ([]domain.BatchRecord, error) {
	query := `INSERT INTO records (url, short_id) VALUES ($1, $2) ON CONFLICT (url) DO UPDATE 
			  SET short_id=records.short_id RETURNING short_id;`
	batch := &pgx.Batch{}
	for i := range data {
		batch.Queue(query, data[i].URL, data[i].ShortID)
	}

	trx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("postgres repository, batch, start trx: %w", err)
	}
	defer trx.Rollback(ctx)

	results := trx.SendBatch(ctx, batch)
	defer results.Close()
	for i := range data {
		err = results.QueryRow().Scan(&data[i].ShortID)
		if err != nil {
			return nil, fmt.Errorf("postgres repository, batch, scan send batch result: %w", err)
		}
	}
	err = trx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres repository, batch, commit trx: %w", err)
	}
	return data, nil
}

func (repo *PostgresRepository) UserURLs(ctx context.Context, uid domain.UID) ([]domain.URLData, error) {
	rows, err := repo.db.Query(ctx, `SELECT url, short_id FROM records
									 JOIN users ON records.id=users.record_id
									 WHERE users.uid=$1;`, uid)
	if err != nil {
		return nil, fmt.Errorf("postgres repository, user urls, query: %w", err)
	}
	result, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.URLData, error) {
		var r domain.URLData
		err := row.Scan(&r.URL, &r.ShortID)
		return r, err
	})
	if err != nil {
		return nil, fmt.Errorf("postgres repository, user urls, collect rows: %w", err)
	}
	return result, nil
}

func (repo *PostgresRepository) Shutdown() error {
	return nil
}
