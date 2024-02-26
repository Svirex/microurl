package server

import (
	"errors"
	"net/http"

	"github.com/Svirex/microurl/internal/apis"
	lg "github.com/Svirex/microurl/internal/logging"
	"github.com/Svirex/microurl/internal/pkg/logging"
	appmiddleware "github.com/Svirex/microurl/internal/pkg/middleware"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Addr string
	API  *apis.ShortenerAPI
}

func NewServer(addr string, baseURL string, generator util.Generator, repository repositories.Repository, shortIDSize uint) *Server {
	return &Server{
		Addr: addr,
		API:  apis.NewShortenerAPI(generator, repository, baseURL, shortIDSize),
	}
}

type Options struct {
	loggingMiddlwareLogger logging.Logger
}

func SetupMiddlewares(router chi.Router, options *Options) {
	router.Use(middleware.Recoverer)
	router.Use(appmiddleware.NewLoggingMiddleware(options.loggingMiddlwareLogger))
}

func (s *Server) SetupRoutes(options *Options) chi.Router {
	router := chi.NewRouter()
	SetupMiddlewares(router, options)

	router.Route("/", apis.GetRoutesFunc(s.API))

	return router
}

func (s *Server) Start() {
	logger, err := lg.NewDefaultLogger()
	if err != nil {
		panic(err)
	}
	options := &Options{
		loggingMiddlwareLogger: logger,
	}
	router := s.SetupRoutes(options)
	err = http.ListenAndServe(s.Addr, router)
	if errors.Is(err, http.ErrServerClosed) {
		return
	} else {
		panic(err)
	}
}
