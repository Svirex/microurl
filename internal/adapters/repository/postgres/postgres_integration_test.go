package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/Svirex/microurl/tests/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	db.Init()
	defer db.Close()

	db.MigrateUp()
	code := m.Run()
	db.MigrateDown()
	fmt.Println("Test end with code ", code)
	os.Exit(code)
}

func setupTest(t *testing.T) (*PostgresRepository, func()) {
	repo := NewPostgresRepository(db.GetPool(), db.GetLogger())

	// tear down later
	return repo, func() {
		// db.GetLogger().Debugln("TEADRDOWN integrations")
		err := db.Truncate()
		require.NoError(t, err)
	}
}

func TestAddGood(t *testing.T) {
	repo, tearDown := setupTest(t)
	defer tearDown()

	data := &domain.Record{
		UID: domain.UID(uuid.New().String()),
		URL: "http://svirex.ru",
	}
	shortID := domain.ShortID("short_id")

	resultShortID, err := repo.Add(context.Background(), shortID, data)
	require.NoError(t, err)
	require.Equal(t, shortID, resultShortID)

	pool := db.GetPool()
	var uid domain.UID
	err = pool.QueryRow(context.Background(), "SELECT uid FROM users WHERE uid=$1", data.UID).Scan(&uid)
	require.NoError(t, err)

}

func TestAddAlreadyExists(t *testing.T) {
	repo, tearDown := setupTest(t)
	defer tearDown()

	data := &domain.Record{
		UID: domain.UID(uuid.New().String()),
		URL: "http://svirex.ru",
	}
	shortID := domain.ShortID("short_id")

	repo.Add(context.Background(), shortID, data)

	shortIDNew := domain.ShortID("short_id_already_exists")

	actualShortID, err := repo.Add(context.Background(), shortIDNew, data)
	require.ErrorIs(t, err, ports.ErrAlreadyExists)
	require.Equal(t, shortID, actualShortID)
}

func TestGetNotFound(t *testing.T) {
	repo, tearDown := setupTest(t)
	defer tearDown()

	_, err := repo.Get(context.Background(), "shotrt_id")
	require.ErrorIs(t, err, ports.ErrNotFound)
}

func TestGetGood(t *testing.T) {
	repo, tearDown := setupTest(t)
	defer tearDown()

	data := &domain.Record{
		UID: domain.UID(uuid.New().String()),
		URL: "http://svirex.ru",
	}
	shortID := domain.ShortID("short_id")

	repo.Add(context.Background(), shortID, data)

	url, err := repo.Get(context.Background(), shortID)
	require.NoError(t, err)
	require.Equal(t, data.URL, url)
}
