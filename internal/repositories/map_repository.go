package repositories

type MapRepository struct {
	data map[string]string
}

var _ Repository = &MapRepository{}

func (m *MapRepository) Add(shortId, url string) error {
	m.data[shortId] = url
	return nil
}

func (m *MapRepository) Get(shortId string) (string, error) {
	return m.data[shortId], nil
}

func NewMapRepository() *MapRepository {
	return &MapRepository{
		data: make(map[string]string),
	}
}
