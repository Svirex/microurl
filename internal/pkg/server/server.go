package server

import (
	"net/http"

	"github.com/Svirex/microurl/internal/pkg/context"
	"github.com/Svirex/microurl/internal/pkg/handlers"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Addr string
}

func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
	}
}

func MainRoutes(appCtx *context.AppContext) chi.Router {
	router := chi.NewRouter()

	router.Route("/", handlers.GetRoutesFunc(appCtx))

	return router
}

func (s *Server) Start(appCtx *context.AppContext) {
	err := http.ListenAndServe(s.Addr, MainRoutes(appCtx))
	if err != nil {
		panic(err)
	}
}
