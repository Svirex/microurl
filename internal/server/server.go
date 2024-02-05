package server

import (
	"fmt"
	"net/http"

	"github.com/Svirex/microurl/internal/handlers"
	"github.com/Svirex/microurl/internal/pkg/context"
)

type Server struct {
	Host string
	Port int
}

func NewServer(host string, port int) *Server {
	return &Server{
		Host: host,
		Port: port,
	}
}

func (s *Server) Start(appCtx *context.AppContext) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.NewMainHandler(appCtx))

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port), mux)
	if err != nil {
		panic(err)
	}
}
