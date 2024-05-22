package inmemory

import (
	"context"
	"fmt"
	"sync"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
)

type ShortenerRepository struct {
	data          map[domain.ShortID]domain.URL
	urlsToShortID map[domain.URL]domain.ShortID
	uidToRecords  map[domain.UID][]domain.URLData
	mutex         sync.Mutex
}

var _ ports.ShortenerRepository = (*ShortenerRepository)(nil)

func NewShortenerRepository() *ShortenerRepository {
	return &ShortenerRepository{
		data:          make(map[domain.ShortID]domain.URL),
		urlsToShortID: make(map[domain.URL]domain.ShortID),
		uidToRecords:  make(map[domain.UID][]domain.URLData),
	}
}

func (m *ShortenerRepository) Add(_ context.Context, shortID domain.ShortID, data *domain.Record) (domain.ShortID, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.addNewOrGetExistShortID(shortID, data.URL, data.UID)
}

func (m *ShortenerRepository) Get(_ context.Context, shortID domain.ShortID) (domain.URL, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	url, ok := m.data[shortID]
	if !ok {
		return domain.URL(""), fmt.Errorf("get url from map repository: %w", ports.ErrNotFound)
	}
	return url, nil
}

func (m *ShortenerRepository) Batch(_ context.Context, uid domain.UID, data []domain.BatchRecord) ([]domain.BatchRecord, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for i := range data {
		record := &data[i]
		shortID, _ := m.addNewOrGetExistShortID(record.ShortID, record.URL, uid)
		record.ShortID = shortID
	}
	return data, nil
}

func (m *ShortenerRepository) UserURLs(ctx context.Context, uid domain.UID) ([]domain.URLData, error) {
	if _, ok := m.uidToRecords[uid]; !ok {
		return make([]domain.URLData, 0), nil
	}
	return m.uidToRecords[uid], nil
}

func (m *ShortenerRepository) Shutdown() error {
	return nil
}

func (m *ShortenerRepository) CheckExists(url domain.URL) (domain.ShortID, bool) {
	shortID, exist := m.urlsToShortID[url]
	return shortID, exist
}

func (m *ShortenerRepository) addNewOrGetExistShortID(shortID domain.ShortID, url domain.URL, uid domain.UID) (domain.ShortID, error) {
	if mapShortID, exist := m.CheckExists(url); exist {
		return mapShortID, fmt.Errorf("add new or get exist short id: %w", ports.ErrAlreadyExists)
	} else {
		m.addNewRecord(shortID, url, uid)
		return shortID, nil
	}
}

func (m *ShortenerRepository) addNewRecord(shortID domain.ShortID, url domain.URL, uid domain.UID) {
	m.data[shortID] = url
	m.urlsToShortID[url] = shortID
	if _, ok := m.uidToRecords[uid]; !ok {
		m.uidToRecords[uid] = make([]domain.URLData, 0)
	}
	m.uidToRecords[uid] = append(m.uidToRecords[uid], domain.URLData{
		ShortID: shortID,
		URL:     url,
	})
}
