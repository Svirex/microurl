package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
)

type MapRepository struct {
	data  map[string]string
	mutex sync.Mutex
}

var _ repositories.URLRepository = (*MapRepository)(nil)

func NewMapRepository() *MapRepository {
	return &MapRepository{
		data: make(map[string]string),
	}
}

func (m *MapRepository) Add(ctx context.Context, d *models.RepositoryAddRecord) (*models.RepositoryGetRecord, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[d.ShortID] = d.URL
	return models.NewRepositoryGetRecord(d.ShortID), nil
}

func (m *MapRepository) Get(ctx context.Context, d *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	u, ok := m.data[d.ShortID]
	if !ok {
		return nil, fmt.Errorf("not found url for %s", d.ShortID)
	}
	return models.NewRepositoryGetResult(u), nil
}

func (m *MapRepository) Shutdown() error {
	return nil
}

func (m *MapRepository) Batch(_ context.Context, batch *models.BatchService) (*models.BatchResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	response := &models.BatchResponse{
		Records: make([]models.BatchResponseRecord, len(batch.Records)),
	}
	for i := range batch.Records {
		m.data[batch.Records[i].ShortURL] = batch.Records[i].URL
		response.Records[i].CorrID = batch.Records[i].CorrID
		response.Records[i].ShortURL = batch.Records[i].ShortURL
	}
	return response, nil
}
