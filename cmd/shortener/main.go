package main

import (
	"time"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/pkg/config"
	"github.com/Svirex/microurl/internal/pkg/context"
	"github.com/Svirex/microurl/internal/server"
	"github.com/Svirex/microurl/internal/storage"
)

func main() {
	config := config.Config{
		Host: "localhost",
		Port: 8080,
	}
	appCtx := context.AppContext{
		Config:     &config,
		Generator:  generators.NewSimpleGenerator(time.Now().UnixNano()),
		Repository: storage.NewMapRepository(),
	}
	server := server.NewServer(config.Host, config.Port)
	server.Start(&appCtx)
}
