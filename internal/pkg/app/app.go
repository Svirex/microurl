package app

import (
	"time"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/pkg/config"
	"github.com/Svirex/microurl/internal/pkg/server"
	"github.com/Svirex/microurl/internal/storage"
)

const SHORT_URL_LENGTH uint = 8

func Run(cfg *config.Config) error {
	generator := generators.NewSimpleGenerator(time.Now().UnixNano())
	repository := storage.NewMapRepository()
	options := server.NewOptions(cfg.Addr, cfg.BaseURL, cfg.FileStoragePath, generator, repository, SHORT_URL_LENGTH)
	server, err := server.NewServer(options)
	if err != nil {
		return err
	}
	return server.Start()
}
