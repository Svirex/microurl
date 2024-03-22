package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/Svirex/microurl/internal/models"
)

type ShortID string
type URL string

type MapRepository struct {
	data          map[ShortID]URL
	urlsToShortID map[URL]ShortID
	mutex         sync.Mutex
}

var _ URLRepository = (*MapRepository)(nil)

func NewMapRepository() *MapRepository {
	return &MapRepository{
		data:          make(map[ShortID]URL),
		urlsToShortID: make(map[URL]ShortID),
	}
}

func (m *MapRepository) Add(ctx context.Context, d *models.RepositoryAddRecord) (*models.RepositoryGetRecord, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if shortID, exist := m.urlsToShortID[URL(d.URL)]; exist {
		return models.NewRepositoryGetRecord(string(shortID)), fmt.Errorf("%w", ErrAlreadyExists)
	} else {
		m.addNewRecord(URL(d.URL), ShortID(d.ShortID))
		return models.NewRepositoryGetRecord(d.ShortID), nil
	}
}

func (m *MapRepository) Get(ctx context.Context, d *models.RepositoryGetRecord) (*models.RepositoryGetResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	u, ok := m.data[ShortID(d.ShortID)]
	if !ok {
		return nil, fmt.Errorf("not found url for %s", d.ShortID)
	}
	return models.NewRepositoryGetResult(string(u)), nil
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
		m.data[ShortID(batch.Records[i].ShortURL)] = URL(batch.Records[i].URL)
		response.Records[i].CorrID = batch.Records[i].CorrID
		response.Records[i].ShortURL = batch.Records[i].ShortURL
	}
	return response, nil
}

func (m *MapRepository) addNewRecord(url URL, shortID ShortID) {
	m.data[shortID] = url
	m.urlsToShortID[url] = shortID
}
