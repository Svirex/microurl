package repositories

import "fmt"

type MapRepository struct {
	data map[string]string
}

var _ Repository = &MapRepository{}

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

func NewMapRepository() *MapRepository {
	return &MapRepository{
		data: make(map[string]string),
	}
}
