package storage

import (
	"fmt"

	"github.com/Svirex/microurl/internal/pkg/repositories"
)

type MapRepository struct {
	data map[string]string
}

var _ repositories.Repository = &MapRepository{}

func (m *MapRepository) Add(shortID, url string) error {
	m.data[shortID] = url
	return nil
}

func (m *MapRepository) Get(shortID string) (*string, error) {
	u, ok := m.data[shortID]
	if !ok {
		return nil, fmt.Errorf("not found url for %s", shortID)
	}
	return &u, nil
}

func NewMapRepository() repositories.Repository {
	return &MapRepository{
		data: make(map[string]string),
	}
}
