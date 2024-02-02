package main

import (
	"time"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/handlers"
	"github.com/Svirex/microurl/internal/repositories"
)

func main() {
	generator := generators.NewSimpleGenerator(time.Now().UnixNano())
	repository := repositories.NewMapRepository()
	server := handlers.NewServer("localhost", 8080, repository, generator)
	server.Start()
}
