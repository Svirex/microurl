package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	bck "github.com/Svirex/microurl/internal/backup"
	"github.com/Svirex/microurl/internal/pkg/backup"
	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/google/uuid"
)

var ErrEmptyFilename = errors.New("empty filename")

var _ repositories.URLRepository = (*FileRepository)(nil)

type FileRepository struct {
	*MapRepository
	backupWriter backup.BackupWriter
}

func NewFileRepository(ctx context.Context, filename string) (*FileRepository, error) {
	if filename == "" {
		return nil, ErrEmptyFilename
	}

	repository := &FileRepository{
		MapRepository: NewMapRepository(),
	}

	var backupWriter backup.BackupWriter

	err := restoreRepository(ctx, filename, repository.MapRepository)
	if err != nil {
		return nil, err
	}

	backupWriter, err = bck.NewFileBackupWriter(filename)
	if err != nil {
		return nil, err
	}

	repository.backupWriter = backupWriter
	return repository, nil
}

func (m *FileRepository) Add(ctx context.Context, d *models.RepositoryAddRecord) error {
	err := m.MapRepository.Add(ctx, d)
	if err != nil {
		return fmt.Errorf("save url to mem storage: %w", err)
	}

	err = m.saveToFile(ctx, d)
	if err != nil {
		return fmt.Errorf("save url to file: %w", err)
	}
	return nil
}

func (m *FileRepository) Get(ctx context.Context, d *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	u, ok := m.data[d.ShortID]
	if !ok {
		return nil, fmt.Errorf("not found url for %s", d.ShortID)
	}
	return models.NewRepositoryGetResult(u), nil
}

func (m *FileRepository) Shutdown() error {
	return nil
}

func restoreRepository(ctx context.Context, filename string, repository *MapRepository) error {
	reader, err := bck.NewFileBackupReader(filename)
	if err != nil {
		return err
	}
	defer reader.Close()
	record, err := reader.Read(ctx)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	for record != nil {
		repository.Add(context.Background(), models.NewRepositoryAddRecord(record.ShortID, record.URL))
		record, err = reader.Read(ctx)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}
	return nil
}

func (m *FileRepository) saveToFile(ctx context.Context, d *models.RepositoryAddRecord) error {
	if m.backupWriter == nil {
		return nil
	}
	record := &backup.Record{
		UUID:    uuid.New().String(),
		ShortID: d.ShortID,
		URL:     d.URL,
	}
	return m.backupWriter.Write(ctx, record)
}
