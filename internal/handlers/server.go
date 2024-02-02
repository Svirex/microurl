package handlers

import (
	"fmt"
	"net/http"

	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/repositories"
)

type Server struct {
	host       string
	port       int
	repository repositories.Repository
	generator  generators.Generator
}

func NewServer(host string, port int, repository repositories.Repository, generator generators.Generator) *Server {
	return &Server{
		host:       host,
		port:       port,
		repository: repository,
		generator:  generator,
	}
}

func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", NewMainHandler(s))

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.host, s.port), mux)
	if err != nil {
		panic(err)
	}
}
