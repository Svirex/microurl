package inmemory

import (
	"context"
	"testing"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/stretchr/testify/require"
)

func TestNewShortenerRepository(t *testing.T) {
	repo := NewShortenerRepository()
	require.Len(t, repo.data, 0)
	require.Len(t, repo.urlsToShortID, 0)
	require.Len(t, repo.uidToRecords, 0)
}

func TestAddNew(t *testing.T) {
	shortID := domain.ShortID("afASDFqwe")
	record := &domain.Record{
		UID: domain.UID("uuid"),
		URL: domain.URL("http://svirex.ru"),
	}
	repo := NewShortenerRepository()
	s, err := repo.Add(context.Background(), shortID, record)
	require.NoError(t, err)
	require.Equal(t, shortID, s)
	url, exists := repo.data[shortID]
	require.True(t, exists)
	require.Equal(t, record.URL, url)
	id, exists := repo.urlsToShortID[record.URL]
	require.True(t, exists)
	require.Equal(t, shortID, id)
	require.Len(t, repo.uidToRecords, 1)
	recs, exists := repo.uidToRecords[record.UID]
	require.True(t, exists)
	require.Equal(t, record.URL, recs[0].URL)
	require.Equal(t, shortID, recs[0].ShortID)
}

func TestAddExisttingRecord(t *testing.T) {
	shortID := domain.ShortID("afASDFqwe")
	record := &domain.Record{
		UID: domain.UID("uuid"),
		URL: domain.URL("http://svirex.ru"),
	}
	repo := NewShortenerRepository()
	s, err := repo.Add(context.Background(), shortID, record)
	require.NoError(t, err)
	require.Equal(t, shortID, s)
	shortID2 := domain.ShortID("gxtye5gsdf")
	s, err = repo.Add(context.Background(), shortID2, record)
	require.ErrorIs(t, err, ports.ErrAlreadyExists)
	require.Equal(t, shortID, s)
}

func TestGetNotFound(t *testing.T) {
	repo := NewShortenerRepository()
	url, err := repo.Get(context.Background(), domain.ShortID("sdfsdfsd"))
	require.ErrorIs(t, err, ports.ErrNotFound)
	require.Equal(t, domain.URL(""), url)
}

func TestGetFound(t *testing.T) {
	shortID := domain.ShortID("afASDFqwe")
	record := &domain.Record{
		UID: domain.UID("uuid"),
		URL: domain.URL("http://svirex.ru"),
	}
	repo := NewShortenerRepository()
	repo.Add(context.Background(), shortID, record)
	url, err := repo.Get(context.Background(), shortID)
	require.NoError(t, err)
	require.Equal(t, record.URL, url)
}

func TestBatch(t *testing.T) {
	batch := make([]domain.BatchRecord, 0, 3)
	batch = append(batch, domain.BatchRecord{
		CorrID:  "1",
		URL:     "http://svirex.ru",
		ShortID: "wrfasd2",
	})
	batch = append(batch, domain.BatchRecord{
		CorrID:  "2",
		URL:     "http://ya.ru",
		ShortID: "q4qwfd",
	})
	batch = append(batch, domain.BatchRecord{
		CorrID:  "3",
		URL:     "http://google.ru",
		ShortID: "436wyefdv",
	})
	repo := NewShortenerRepository()
	b, err := repo.Batch(context.Background(), domain.UID(""), batch)
	require.NoError(t, err)
	require.Equal(t, batch, b)
}

func TestBatchWithExist(t *testing.T) {
	shortID := domain.ShortID("afASDFqwe")
	record := &domain.Record{
		UID: domain.UID("uuid"),
		URL: domain.URL("http://svirex.ru"),
	}
	repo := NewShortenerRepository()
	repo.Add(context.Background(), shortID, record)
	batch := make([]domain.BatchRecord, 0, 3)
	batch = append(batch, domain.BatchRecord{
		CorrID:  "1",
		URL:     "http://svirex.ru",
		ShortID: "wrfasd2",
	})
	batch = append(batch, domain.BatchRecord{
		CorrID:  "2",
		URL:     "http://ya.ru",
		ShortID: "q4qwfd",
	})
	batch = append(batch, domain.BatchRecord{
		CorrID:  "3",
		URL:     "http://google.ru",
		ShortID: "436wyefdv",
	})
	b, err := repo.Batch(context.Background(), domain.UID(""), batch)
	require.NoError(t, err)
	require.Len(t, b, 3)
	require.Equal(t, shortID, b[0].ShortID)
	require.Equal(t, record.URL, b[0].URL)
}
