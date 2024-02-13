package server

import (
	"errors"
	"net/http"

	"github.com/Svirex/microurl/internal/apis"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Addr string
	API  *apis.ShortenerApi
}

func NewServer(addr string, baseURL string, generator util.Generator, repository repositories.Repository, shortIDSize uint) *Server {
	return &Server{
		Addr: addr,
		API:  apis.NewShortenerApi(generator, repository, baseURL, shortIDSize),
	}
}

func (s *Server) SetupRoutes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	router.Route("/", apis.GetRoutesFunc(s.API))

	return router
}

func (s *Server) Start() {
	router := s.SetupRoutes()
	err := http.ListenAndServe(s.Addr, router)
	if errors.Is(err, http.ErrServerClosed) {
		return
	} else {
		panic(err)
	}
}
