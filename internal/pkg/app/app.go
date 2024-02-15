package app

import (
	"time"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/pkg/server"
	"github.com/Svirex/microurl/internal/storage"
)

func Run(addr string, baseURL string) {
	generator := generators.NewSimpleGenerator(time.Now().UnixNano())
	repository := storage.NewMapRepository()
	server := server.NewServer(addr, baseURL, generator, repository, 8)
	server.Start()
}
