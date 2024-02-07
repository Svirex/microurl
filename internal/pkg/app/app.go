package app

import (
	"time"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/pkg/config"
	"github.com/Svirex/microurl/internal/pkg/context"
	"github.com/Svirex/microurl/internal/pkg/server"
	"github.com/Svirex/microurl/internal/storage"
)

func Run(addr string, baseURL string) {
	config := config.Config{
		Addr:    addr,
		BaseURL: baseURL,
	}
	appCtx := context.AppContext{
		Config:     &config,
		Generator:  generators.NewSimpleGenerator(time.Now().UnixNano()),
		Repository: storage.NewMapRepository(),
	}
	server := server.NewServer(addr)
	server.Start(&appCtx)
}
