package server

import (
	"context"
	"net"
	"net/http"
)

// type Server struct {
// 	Addr string
// 	API  *apis.ShortenerAPI
// }

func NewServer(ctx context.Context, addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:        addr,
		Handler:     handler,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
}

// func NewServer(addr, baseURL string, generator util.Generator, repository repositories.Repository, shortIDSize uint) (*Server, error) {
// 	shortenerAPI, err := apis.NewShortenerAPI(baseURL, generator, repository, shortIDSize)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Server{
// 		Addr: addr,
// 		API:  shortenerAPI,
// 	}, nil
// }

// type options struct {
// 	loggingMiddlwareLogger logging.Logger
// }

// func SetupMiddlewares(router chi.Router, options *options) {
// 	router.Use(middleware.Recoverer)
// 	router.Use(appmiddleware.NewLoggingMiddleware(options.loggingMiddlwareLogger))
// 	router.Use(appmiddleware.GzipHandler)
// 	router.Use(middleware.Compress(5, "text/html", "application/json"))
// }

// func (s *Server) SetupRoutes(options *options) chi.Router {
// 	router := chi.NewRouter()
// 	SetupMiddlewares(router, options)

// 	router.Route("/", apis.GetRoutesFunc(s.API))

// 	return router
// }

// func (s *Server) Start() error {
// 	logger, err := lg.NewDefaultLogger()
// 	if err != nil {
// 		return err
// 	}
// 	options := &options{
// 		loggingMiddlwareLogger: logger,
// 	}
// 	router := s.SetupRoutes(options)
// 	err = http.ListenAndServe(s.Addr, router)
// 	if errors.Is(err, http.ErrServerClosed) {
// 		return nil
// 	} else {
// 		return err
// 	}
// }
