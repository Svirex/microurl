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
	err := repo.writeToFile(backupRecord)
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

func (repo *ShortenerRepository) Get(ctx context.Context, shortID domain.ShortID) (domain.URL, error) {
	return repo.repo.Get(ctx, shortID)
}

func (repo *ShortenerRepository) Batch(ctx context.Context, uid domain.UID, data []domain.BatchRecord) ([]domain.BatchRecord, error) {
	backupRecords := make([]domain.BackupRecord, 0, len(data))
	for i := range data {
		record := &data[i]
		backupRecords = append(backupRecords, domain.BackupRecord{
			UUID:    uuid.New().String(),
			ShortID: record.ShortID,
			URL:     record.URL,
			UID:     uid,
		})
	}
	err := repo.writeBatchToFile(backupRecords)
	if err != nil {
		return nil, fmt.Errorf("file repository, batch, write to file: %w", err)
	}
	return repo.repo.Batch(ctx, uid, data)
}

func (repo *ShortenerRepository) writeBatchToFile(data []domain.BackupRecord) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for i := range data {
		record := &data[i]
		err := repo.writer.Write(context.Background(), record)
		if err != nil {
			return fmt.Errorf("write batch to file: %w", err)
		}
	}
	return nil
}

func (repo *ShortenerRepository) UserURLs(ctx context.Context, uid domain.UID) ([]domain.URLData, error) {
	return repo.repo.UserURLs(ctx, uid)
}

func (repo *ShortenerRepository) Shutdown() error {
	return nil
}
