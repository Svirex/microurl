package service

import (
	"context"
	"testing"

	"github.com/Svirex/microurl/internal/adapters/generator"
	"github.com/Svirex/microurl/internal/adapters/repository/inmemory"
	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/google/uuid"
)

func setupBenchmarkTest(name string) (*ShortenerService, func()) {
	repo := inmemory.NewShortenerRepository()
	gen := generator.NewStringGenerator(255)
	service := NewShortenerService(gen, repo, 8, "http://localhost:8090")
	// fmt.Println("[TEST] ", name)

	// tear down later
	return service, func() {
		// db.GetLogger().Debugln("TEADRDOWN benchmark")
		// fmt.Println("[TEST ENDED]")
	}
}

func BenchmarkAddGood(b *testing.B) {
	service, tearDown := setupBenchmarkTest("BenchmarkAddGood")
	defer tearDown()
	data := &domain.Record{
		UID: domain.UID(uuid.New().String()),
		URL: "http://svirex.ru",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.Add(context.Background(), data)
	}
}
