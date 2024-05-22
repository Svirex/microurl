package file

import (
	"context"
	"fmt"
	"sync"

	"github.com/Svirex/microurl/internal/adapters/repository/inmemory"
	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/google/uuid"
)

type ShortenerRepository struct {
	repo   *inmemory.ShortenerRepository
	writer ports.BackupWriter
	mutex  sync.Mutex
}

func NewShortenerRepository(inmemoryRepo *inmemory.ShortenerRepository, writer ports.BackupWriter) *ShortenerRepository {
	return &ShortenerRepository{
		repo:   inmemoryRepo,
		writer: writer,
	}
}

var _ ports.ShortenerRepository = (*ShortenerRepository)(nil)

func (repo *ShortenerRepository) Add(ctx context.Context, shortID domain.ShortID, data *domain.Record) (domain.ShortID, error) {
	if id, exist := repo.repo.CheckExists(data.URL); exist {
		return id, fmt.Errorf("file repository, add: %w", ports.ErrAlreadyExists)
	}
	backupRecord := &domain.BackupRecord{
		UUID:    uuid.New().String(),
		ShortID: shortID,
		URL:     data.URL,
		UID:     data.UID,
	}
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	err := repo.writer.Write(ctx, backupRecord)
	if err != nil {
		return domain.ShortID(""), fmt.Errorf("file repository, add, write to file: %w", err)
	}
	repo.repo.Add(ctx, shortID, data)
	return shortID, nil
}

func (repo *ShortenerRepository) writeToFile(record *domain.BackupRecord) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	return repo.writer.Write(context.Background(), record)
}
