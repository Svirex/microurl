package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/Svirex/microurl/internal/backup"
	"github.com/Svirex/microurl/internal/models"
	"github.com/google/uuid"
)

var ErrEmptyFilename = errors.New("empty filename")

var _ URLRepository = (*FileRepository)(nil)

type FileRepository struct {
	*MapRepository
	backupWriter backup.BackupWriter
	mutex        sync.Mutex
}

func NewFileRepository(ctx context.Context, filename string) (*FileRepository, error) {
	if filename == "" {
		return nil, fmt.Errorf("create new file repository: %w", ErrEmptyFilename)
	}

	repository := &FileRepository{
		MapRepository: NewMapRepository(),
	}

	var backupWriter backup.BackupWriter

	err := restoreRepository(ctx, filename, repository.MapRepository)
	if err != nil {
		return nil, fmt.Errorf("create new file repository, restore data: %w", err)
	}

	backupWriter, err = backup.NewFileBackupWriter(filename)
	if err != nil {
		return nil, fmt.Errorf("create new file repository, new backup writer: %w", err)
	}

	repository.backupWriter = backupWriter
	return repository, nil
}

func (m *FileRepository) Add(ctx context.Context, d *models.RepositoryAddRecord) (*models.RepositoryGetRecord, error) {
	res, err := m.MapRepository.Add(ctx, d)
	if errors.Is(err, ErrAlreadyExists) {
		return res, fmt.Errorf("short_id for url in MapRepository already exist: %w", err)
	} else if err != nil {
		return nil, fmt.Errorf("save url to mem storage: %w", err)
	}

	err = m.saveToFile(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("save url to file: %w", err)
	}
	return res, nil
}

func (m *FileRepository) Get(ctx context.Context, d *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	return m.MapRepository.Get(ctx, d)
}

func (m *FileRepository) Shutdown() error {
	return nil
}

func restoreRepository(ctx context.Context, filename string, repository *MapRepository) error {
	reader, err := backup.NewFileBackupReader(filename)
	if err != nil {
		return fmt.Errorf("restore data, new backup reader: %w", err)
	}
	defer reader.Close()
	record, err := reader.Read(ctx)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("restore data, read: %w", err)
	}
	for record != nil {
		repository.Add(context.Background(), models.NewRepositoryAddRecord(record.ShortID, record.URL))
		record, err = reader.Read(ctx)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("restore data, while read: %w", err)
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
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.backupWriter.Write(ctx, record)
}

func (m *FileRepository) saveBatchToFile(ctx context.Context, batch *models.BatchService) error {
	if m.backupWriter == nil {
		return nil
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for i := range batch.Records {
		record := &backup.Record{
			UUID:    uuid.New().String(),
			ShortID: batch.Records[i].ShortURL,
			URL:     batch.Records[i].URL,
		}
		err := m.backupWriter.Write(ctx, record)
		if err != nil {
			return fmt.Errorf("error save url to file: %w", err)
		}
	}

	return nil
}

func (m *FileRepository) Batch(ctx context.Context, batch *models.BatchService) (*models.BatchResponse, error) {
	_, err := m.MapRepository.Batch(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("save url to mem storage: %w", err)
	}

	err = m.saveBatchToFile(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("save url to file: %w", err)
	}
	response := &models.BatchResponse{
		Records: make([]models.BatchResponseRecord, len(batch.Records)),
	}
	for i := range batch.Records {
		response.Records[i].CorrID = batch.Records[i].CorrID
		response.Records[i].ShortURL = batch.Records[i].ShortURL
	}
	return response, nil
}

func (m *FileRepository) UserURLs(_ context.Context, uid string) ([]models.UserURLRecord, error) {
	result := make([]models.UserURLRecord, 0)

	return result, nil
}
