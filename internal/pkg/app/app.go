package app

import (
	"time"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/pkg/config"
	"github.com/Svirex/microurl/internal/pkg/server"
	"github.com/Svirex/microurl/internal/storage"
)

const shortURLLength uint = 8

func Run(cfg *config.Config) error {
	generator := generators.NewSimpleGenerator(time.Now().UnixNano())
	repository, err := storage.NewMapRepository(cfg.FileStoragePath)
	if err != nil {
		return err
	}
	server, err := server.NewServer(cfg.Addr, cfg.BaseURL, generator, repository, shortURLLength)
	if err != nil {
		return err
	}
	return server.Start()
}
