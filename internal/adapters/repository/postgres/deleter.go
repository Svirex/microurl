package postgres

import (
	"context"
	"fmt"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeleterRepository - репозиторий.
type DeleterRepository struct {
	db     *pgxpool.Pool
	logger ports.Logger
}

// NewDeleterRepository - новый репозиторий
func NewDeleterRepository(db *pgxpool.Pool, logger ports.Logger) *DeleterRepository {
	return &DeleterRepository{
		db:     db,
		logger: logger,
	}
}

var _ ports.DeleterRepository = (*DeleterRepository)(nil)

// Delete - помечает урлы как удаленные
func (r *DeleterRepository) Delete(ctx context.Context, batch []*domain.DeleteData) error {
	uids := make([]string, 0)
	shortIDs := make([]string, 0)
	for _, v := range batch {
		if v != nil {
			uids = append(uids, v.UID)
			shortIDs = append(shortIDs, v.ShortID)
		}
	}
	i := 1
	uidsPlacement := ""
	first := true
	for ; i <= len(uids); i++ {
		if first {
			uidsPlacement += fmt.Sprintf("$%d", i)
			first = false
		} else {
			uidsPlacement += fmt.Sprintf(",$%d", i)
		}

	}
	first = true
	shortIDsPlacement := ""
	for ; i <= 2*len(shortIDs); i++ {
		if first {
			shortIDsPlacement += fmt.Sprintf("$%d", i)
			first = false
		} else {
			shortIDsPlacement += fmt.Sprintf(",$%d", i)
		}
	}
	values := make([]interface{}, 0, 2*len(uids))
	for _, v := range uids {
		values = append(values, v)
	}
	for _, v := range shortIDs {
		values = append(values, v)
	}

	_, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE records SET is_deleted=true
				FROM (
					SELECT records.id FROM records
					JOIN users ON records.id=users.record_id
					WHERE users.uid IN (%s) AND records.short_id IN (%s)
				) as d
				WHERE records.id=d.id;`, uidsPlacement, shortIDsPlacement), values...)
	if err != nil {
		return fmt.Errorf("deleter repository, delete: %w", err)
	}
	return nil
}
