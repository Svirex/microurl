package ports

import (
	"context"
	"errors"

	"github.com/Svirex/microurl/internal/core/domain"
)

var ErrAlreadyExists = errors.New("already exists")
var ErrNotFound = errors.New("not found")

type ShortenerService interface {
	Add(context.Context, *domain.Record) (domain.ShortURL, error)
	Get(ctx context.Context, shortID domain.ShortID) (domain.URL, error)
	Batch(ctx context.Context, uid domain.UID, data []domain.BatchRecord) ([]domain.BatchRecord, error)
	UserURLs(ctx context.Context, uid domain.UID) ([]domain.URLData, error)
	Shutdown() error
}

type DBCheckerService interface {
	Ping(context.Context) error
	Shutdown() error
}

type ShortenerRepository interface {
	Add(ctx context.Context, shortID domain.ShortID, data *domain.Record) (domain.ShortID, error)
	Get(ctx context.Context, shortID domain.ShortID) (domain.URL, error)
	Batch(ctx context.Context, uid domain.UID, data []domain.BatchRecord) ([]domain.BatchRecord, error)
	UserURLs(ctx context.Context, uid domain.UID) ([]domain.URLData, error)
	Shutdown() error
}

type BackupWriter interface {
	Write(ctx context.Context, record *domain.BackupRecord) error
}

type BackupReader interface {
	Next() bool
	Read(ctx context.Context) (*domain.BackupRecord, error)
}

type StringGenerator interface {
	Generate(ctx context.Context, size uint) string
}

type DeleterService interface {
	Process(ctx context.Context, uid string, shortIDs []string)
	Run() error
	Shutdown() error
}

type DeleterRepository interface {
	Delete(ctx context.Context, batch []*domain.DeleteData) error
}

type DBCheck interface {
	Ping(context.Context) error
}
