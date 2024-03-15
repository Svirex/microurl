package services

import (
	"context"
	"errors"

	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/pkg/services"
	"github.com/Svirex/microurl/internal/pkg/util"
)

var ErrUnableAddRecord = errors.New("unable add record into repository")
var ErrNotFound = errors.New("record not found")
var ErrSomethingWrong = errors.New("unknown error")
var ErrUnableBackupRecord = errors.New("unable write record into backup")

type ShortenerService struct {
	Generator   util.Generator
	Repository  repositories.URLRepository
	ShortIDSize uint
}

func NewShortenerService(generator util.Generator, repository repositories.URLRepository, shortIDSize uint) *ShortenerService {
	return &ShortenerService{
		Generator:   generator,
		Repository:  repository,
		ShortIDSize: shortIDSize,
	}
}

var _ services.Shortener = (*ShortenerService)(nil)

func (s *ShortenerService) Add(ctx context.Context, d *models.ServiceAddRecord) (*models.ServiceAddResult, error) {
	shortID := s.generateShortID()
	res, err := s.Repository.Add(ctx, models.NewRepositoryAddRecord(shortID, d.URL))
	if err != nil {
		return nil, ErrUnableAddRecord
	}
	return models.NewServiceAddResult(res.ShortID), nil
}

func (s *ShortenerService) Get(ctx context.Context, d *models.ServiceGetRecord) (*models.ServiceGetResult, error) {
	result, err := s.Repository.Get(ctx, models.NewRepositoryGetRecord(d.ShortID))
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

func (s *ShortenerService) Shutdown() error {
	return nil
}
