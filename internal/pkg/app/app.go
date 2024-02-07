package app

import (
	"time"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/pkg/config"
	"github.com/Svirex/microurl/internal/pkg/context"
	"github.com/Svirex/microurl/internal/pkg/server"
	"github.com/Svirex/microurl/internal/storage"
)

func Run(host string, port int) {
	config := config.Config{
		Host: host,
		Port: port,
	}
	appCtx := context.AppContext{
		Config:     &config,
		Generator:  generators.NewSimpleGenerator(time.Now().UnixNano()),
		Repository: storage.NewMapRepository(),
	}
	server := server.NewServer(config.Host, config.Port)
	server.Start(&appCtx)
}
