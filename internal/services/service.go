package services

import (
	"errors"

	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/pkg/services"
	"github.com/Svirex/microurl/internal/pkg/util"
)

var ErrUnableAddRecord = errors.New("unable add record into repository")
var ErrNotFound = errors.New("record not found")
var ErrSomethingWrong = errors.New("unknown error")

type ShortenerService struct {
	Generator   util.Generator
	Repository  repositories.Repository
	ShortIDSize uint
}

func NewShortenerService(generator util.Generator, repository repositories.Repository, shortIDSize uint) services.Shortener {
	return &ShortenerService{
		Generator:   generator,
		Repository:  repository,
		ShortIDSize: shortIDSize,
	}
}

var _ services.Shortener = (*ShortenerService)(nil)

func (s *ShortenerService) Add(d *models.ServiceAddRecord) (*models.ServiceAddResult, error) {
	shortID := s.generateShortID()
	err := s.Repository.Add(models.NewRepositoryAddRecord(shortID, d.URL))
	if err != nil {
		return nil, ErrUnableAddRecord
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
