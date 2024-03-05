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

func NewServer(options *Options) (*Server, error) {
	apiOptions := apis.NewOptions(options.BaseURL, options.FileBackupPath, options.Generator, options.Repository, options.ShortIDSize)
	shortenerAPI, err := apis.NewShortenerAPI(apiOptions)
	if err != nil {
		return nil, err
	}
	return &Server{
		Addr: options.Addr,
		API:  shortenerAPI,
	}, nil
}

type options struct {
	loggingMiddlwareLogger logging.Logger
}

func SetupMiddlewares(router chi.Router, options *options) {
	router.Use(middleware.Recoverer)
	router.Use(appmiddleware.NewLoggingMiddleware(options.loggingMiddlwareLogger))
	router.Use(appmiddleware.GzipHandler)
	router.Use(middleware.Compress(5, "text/html", "application/json"))
}

func (s *Server) SetupRoutes(options *options) chi.Router {
	router := chi.NewRouter()
	SetupMiddlewares(router, options)

	router.Route("/", apis.GetRoutesFunc(s.API))

	return router
}

func (s *Server) Start() error {
	logger, err := lg.NewDefaultLogger()
	if err != nil {
		return err
	}
	options := &options{
		loggingMiddlwareLogger: logger,
	}
	router := s.SetupRoutes(options)
	err = http.ListenAndServe(s.Addr, router)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return err
	}
}

type Options struct {
	Addr           string
	BaseURL        string
	FileBackupPath string
	Generator      util.Generator
	Repository     repositories.Repository
	ShortIDSize    uint
}

func NewOptions(addr, baseURL, fileBackupPath string, generator util.Generator, repository repositories.Repository, shortIDSize uint) *Options {
	return &Options{
		Addr:           addr,
		BaseURL:        baseURL,
		FileBackupPath: fileBackupPath,
		Generator:      generator,
		Repository:     repository,
		ShortIDSize:    shortIDSize,
	}
}
