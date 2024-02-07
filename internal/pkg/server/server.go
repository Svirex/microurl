package server

import (
	"fmt"
	"net/http"

	"github.com/Svirex/microurl/internal/pkg/context"
	"github.com/Svirex/microurl/internal/pkg/handlers"
	"github.com/go-chi/chi/v5"
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

func MainRoutes(appCtx *context.AppContext) chi.Router {
	router := chi.NewRouter()

	router.Route("/", handlers.GetRoutesFunc(appCtx))

	return router
}

func (s *Server) Start(appCtx *context.AppContext) {
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port), MainRoutes(appCtx))
	if err != nil {
		panic(err)
	}
}
