package storage

import (
	"fmt"
	"sync"

	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
)

type MapRepository struct {
	data  map[string]string
	mutex sync.Mutex
}

var _ repositories.Repository = (*MapRepository)(nil)

func (m *MapRepository) Add(d *models.RepositoryAddRecord) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[d.ShortID] = d.URL
	return nil
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

func NewMapRepository() repositories.Repository {
	return &MapRepository{
		data: make(map[string]string),
	}
}
