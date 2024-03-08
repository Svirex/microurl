package storage

import (
	"errors"
	"fmt"
	"io"
	"sync"

	bck "github.com/Svirex/microurl/internal/backup"
	"github.com/Svirex/microurl/internal/pkg/backup"
	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/google/uuid"
)

type MapRepository struct {
	data         map[string]string
	mutex        sync.Mutex
	backupWriter backup.BackupWriter
}

var _ repositories.Repository = (*MapRepository)(nil)

func (m *MapRepository) Add(d *models.RepositoryAddRecord) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[d.ShortID] = d.URL
	return m.backup(d)
}

func (m *MapRepository) Get(d *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	u, ok := m.data[d.ShortID]
	if !ok {
		return nil, fmt.Errorf("not found url for %s", d.ShortID)
	}
	return models.NewRepositoryGetResult(u), nil
}

func NewMapRepository(filename string) (repositories.Repository, error) {
	repository := &MapRepository{
		data: make(map[string]string),
	}
	var backupWriter backup.BackupWriter
	if filename != "" {
		err := restoreRepository(filename, repository)
		if err != nil {
			return nil, err
		}
		backupWriter, err = bck.NewFileBackupWriter(filename)
		if err != nil {
			return nil, err
		}
	}
	repository.backupWriter = backupWriter
	return repository, nil
}
func restoreRepository(filename string, repository repositories.Repository) error {
	if filename == "" {
		return errors.New("filename is empty")
	}
	reader, err := bck.NewFileBackupReader(filename)
	if err != nil {
		return err
	}
	defer reader.Close()
	record, err := reader.Read()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	for record != nil {
		repository.Add(models.NewRepositoryAddRecord(record.ShortID, record.URL))
		record, err = reader.Read()
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}
	return nil
}

func (m *MapRepository) backup(d *models.RepositoryAddRecord) error {
	if m.backupWriter == nil {
		return nil
	}
	record := &backup.Record{
		UUID:    uuid.New().String(),
		ShortID: d.ShortID,
		URL:     d.URL,
	}
	return m.backupWriter.Write(record)
}
