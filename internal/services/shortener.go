package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/models"
	"github.com/Svirex/microurl/internal/storage"
)

type Shortener interface {
	Add(context.Context, *models.ServiceAddRecord) (*models.ServiceAddResult, error)
	Get(context.Context, *models.ServiceGetRecord) (*models.ServiceGetResult, error)
	Batch(context.Context, *models.BatchRequest) (*models.BatchResponse, error)
	UserURLs(context.Context, string) ([]models.UserURLRecord, error)
	Shutdown() error
}

var ErrNotFound = errors.New("record not found")
var ErrSomethingWrong = errors.New("unknown error")
var ErrUnableBackupRecord = errors.New("unable write record into backup")

type ShortenerService struct {
	Generator   generators.Generator
	Repository  storage.URLRepository
	ShortIDSize uint
}

func NewShortenerService(generator generators.Generator, repository storage.URLRepository, shortIDSize uint) *ShortenerService {
	return &ShortenerService{
		Generator:   generator,
		Repository:  repository,
		ShortIDSize: shortIDSize,
	}
}

var _ Shortener = (*ShortenerService)(nil)

func (s *ShortenerService) Add(ctx context.Context, d *models.ServiceAddRecord) (*models.ServiceAddResult, error) {
	shortID := s.generateShortID()
	res, err := s.Repository.Add(ctx, models.NewRepositoryAddRecord(shortID, d.URL, d.UID))
	if errors.Is(err, storage.ErrAlreadyExists) {
		return models.NewServiceAddResult(res.ShortID), fmt.Errorf("short id for url already exist: %w", err)
	} else if err != nil {
		return nil, fmt.Errorf("repository add: %w", err)
	}
	return models.NewServiceAddResult(res.ShortID), nil
}

func (s *ShortenerService) Get(ctx context.Context, d *models.ServiceGetRecord) (*models.ServiceGetResult, error) {
	result, err := s.Repository.Get(ctx, models.NewRepositoryGetRecord(d.ShortID))
	if errors.Is(err, storage.ErrNotFound) {
		return nil, fmt.Errorf("%w: service get by short id: %w", ErrNotFound, err)
	} else if err != nil {
		return nil, fmt.Errorf("service get by short id: %w", err)
	}
	return models.NewServiceGetResult(result.URL), nil
}

func (s *ShortenerService) generateShortID() string {
	return s.Generator.RandString(s.ShortIDSize)
}

func (s *ShortenerService) Shutdown() error {
	return nil
}

func (s *ShortenerService) Batch(ctx context.Context, batch *models.BatchRequest) (*models.BatchResponse, error) {
	batchService := &models.BatchService{
		Records: make([]models.BatchServiceRecord, len(batch.Records)),
	}
	for i := range batch.Records {
		batchService.Records[i].CorrID = batch.Records[i].CorrID
		batchService.Records[i].URL = batch.Records[i].URL
		batchService.Records[i].ShortURL = s.generateShortID()
	}
	result, err := s.Repository.Batch(ctx, batchService)
	if err != nil {
		return nil, fmt.Errorf("service batch add: %w", err)
	}
	return result, nil
}

func (s *ShortenerService) UserURLs(ctx context.Context, uid string) ([]models.UserURLRecord, error) {
	result, err := s.Repository.UserURLs(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("service get user ulrs: %w", err)
	}
	return result, nil
}
