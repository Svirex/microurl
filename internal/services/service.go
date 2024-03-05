package services

import (
	"errors"

	bck "github.com/Svirex/microurl/internal/backup"
	"github.com/Svirex/microurl/internal/pkg/backup"
	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/pkg/services"
	"github.com/Svirex/microurl/internal/pkg/util"
	"github.com/google/uuid"
)

var ErrUnableAddRecord = errors.New("unable add record into repository")
var ErrNotFound = errors.New("record not found")
var ErrSomethingWrong = errors.New("unknown error")
var ErrUnableBackupRecord = errors.New("unable write record into backup")

type ShortenerService struct {
	Generator    util.Generator
	Repository   repositories.Repository
	ShortIDSize  uint
	backupWriter backup.BackupWriter
}

type Options struct {
	Generator      util.Generator
	Repository     repositories.Repository
	ShortIDSize    uint
	FileBackupPath string
}

func NewOptions(fileBackupPath string, generator util.Generator, repository repositories.Repository, shortIDSize uint) *Options {
	return &Options{
		Generator:      generator,
		Repository:     repository,
		ShortIDSize:    shortIDSize,
		FileBackupPath: fileBackupPath,
	}
}

func NewShortenerService(options *Options) (services.Shortener, error) {
	var backupWriter backup.BackupWriter
	if options.FileBackupPath != "" {
		err := restoreRepository(options.FileBackupPath, options.Repository)
		if err != nil {
			return nil, err
		}
		backupWriter, err = bck.NewFileBackupWriter(options.FileBackupPath)
		if err != nil {
			return nil, err
		}
	}
	return &ShortenerService{
		Generator:    options.Generator,
		Repository:   options.Repository,
		ShortIDSize:  options.ShortIDSize,
		backupWriter: backupWriter,
	}, nil
}

var _ services.Shortener = (*ShortenerService)(nil)

func (s *ShortenerService) Add(d *models.ServiceAddRecord) (*models.ServiceAddResult, error) {
	shortID := s.generateShortID()
	err := s.Repository.Add(models.NewRepositoryAddRecord(shortID, d.URL))
	if err != nil {
		return nil, ErrUnableAddRecord
	}
	err = s.backup(shortID, d.URL)
	if err != nil {
		return nil, ErrUnableBackupRecord
	}
	return models.NewServiceAddResult(shortID), nil
}

func (s *ShortenerService) Get(d *models.ServiceGetRecord) (*models.ServiceGetResult, error) {
	result, err := s.Repository.Get(models.NewRepositoryGetRecord(d.ShortID))
	if errors.Is(err, repositories.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, ErrSomethingWrong
	}
	return models.NewServiceGetResult(result.URL), nil
}

func (s *ShortenerService) generateShortID() string {
	return s.Generator.RandString(s.ShortIDSize)
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
	if err != nil {
		return err
	}
	for record != nil {
		repository.Add(models.NewRepositoryAddRecord(record.ShortID, record.URL))
		record, err = reader.Read()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ShortenerService) backup(shortID, URL string) error {
	if s.backupWriter == nil {
		return nil
	}
	record := &backup.Record{
		UUID:    uuid.New().String(),
		ShortID: shortID,
		URL:     URL,
	}
	return s.backupWriter.Write(record)
}
