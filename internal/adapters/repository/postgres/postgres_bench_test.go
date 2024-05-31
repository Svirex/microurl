//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/tests/db"
	"github.com/google/uuid"
)

func setupBenchmarkTest() (*PostgresRepository, func()) {
	repo := NewPostgresRepository(db.GetPool(), db.GetLogger())

	// tear down later
	return repo, func() {
		// db.GetLogger().Debugln("TEADRDOWN benchmark")
		db.Truncate()
	}
}

func BenchmarkAddGood(b *testing.B) {
	repo, tearDown := setupBenchmarkTest()
	defer tearDown()
	data := &domain.Record{
		UID: domain.UID(uuid.New().String()),
		URL: "http://svirex.ru",
	}
	shortID := domain.ShortID("short_id")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		repo.Add(context.Background(), shortID, data)
		b.StopTimer()
		db.Truncate()
		b.StartTimer()
	}
}
