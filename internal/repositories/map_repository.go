package repositories

import "fmt"

type MapRepository struct {
	data map[string]string
}

var _ Repository = &MapRepository{}

func (m *MapRepository) Add(shortId, url string) error {
	m.data[shortId] = url
	return nil
}

func (m *MapRepository) Get(shortId string) (*string, error) {
	u, ok := m.data[shortId]
	if !ok {
		return nil, fmt.Errorf("Not found url for %s", shortId)
	}
	return &u, nil
}

func NewMapRepository() *MapRepository {
	return &MapRepository{
		data: make(map[string]string),
	}
}
