package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
)

type ShortenerService struct {
	shortIDGenerator ports.StringGenerator
	repository       ports.ShortenerRepository
	shortIDSize      uint
	baseURL          string
}

func NewShortenerService(
	shortIDGenerator ports.StringGenerator,
	repository ports.ShortenerRepository,
	shortIDSize uint,
	baseURL string,
) *ShortenerService {
	return &ShortenerService{
		shortIDGenerator: shortIDGenerator,
		repository:       repository,
		shortIDSize:      shortIDSize,
		baseURL:          baseURL,
	}
}

var _ ports.ShortenerService = (*ShortenerService)(nil)

func (s *ShortenerService) Add(ctx context.Context, record *domain.Record) (domain.ShortURL, error) {
	shortID := domain.ShortID(s.shortIDGenerator.Generate(ctx, s.shortIDSize))
	id, err := s.repository.Add(ctx, shortID, record)
	if err != nil {
		if errors.Is(err, ports.ErrAlreadyExists) {
			return s.shortURL(id), err
		}
		return domain.ShortURL(""), fmt.Errorf("shortener service, add: %w", err)
	}
	return s.shortURL(id), nil
}

func (s *ShortenerService) Get(ctx context.Context, shortID domain.ShortID) (domain.URL, error) {
	return s.repository.Get(ctx, shortID)
}

func (s *ShortenerService) Batch(ctx context.Context, uid domain.UID, data []domain.BatchRecord) ([]domain.BatchRecord, error) {
	for i := range data {
		data[i].ShortID = domain.ShortID(s.shortIDGenerator.Generate(ctx, s.shortIDSize))
	}
	data, err := s.repository.Batch(ctx, uid, data)
	if err != nil {
		return nil, fmt.Errorf("shortener service, batch: %w", err)
	}
	for i := range data {
		data[i].ShortURL = domain.URL(s.shortURL(data[i].ShortID))
	}
	return data, nil
}

func (s *ShortenerService) UserURLs(ctx context.Context, uid domain.UID) ([]domain.URLData, error) {
	data, err := s.repository.UserURLs(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("shortener service, user url: %w", err)
	}
	for i := range data {
		data[i].ShortURL = domain.URL(s.shortURL(data[i].ShortID))
	}
	return data, nil
}

func (s *ShortenerService) Shutdown() error {
	return nil
}

func (s *ShortenerService) shortURL(shortID domain.ShortID) domain.ShortURL {
	return domain.ShortURL(fmt.Sprintf("%s/%s", s.baseURL, string(shortID)))
}
