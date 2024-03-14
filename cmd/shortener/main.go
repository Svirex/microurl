package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Svirex/microurl/internal/apis"
	"github.com/Svirex/microurl/internal/generators"
	"github.com/Svirex/microurl/internal/logging"
	"github.com/Svirex/microurl/internal/pkg/config"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/pkg/server"
	srv "github.com/Svirex/microurl/internal/pkg/services"
	"github.com/Svirex/microurl/internal/services"
	"github.com/Svirex/microurl/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const shortURLLength uint = 8

func main() {
	cfg := config.Parse()

	logger, err := logging.NewDefaultLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Shutdown()

	generator := generators.NewSimpleGenerator(time.Now().UnixNano())
	logger.Info("Created generator...")

	var repository repositories.URLRepository

	serverCtx, serverCancel := context.WithCancel(context.Background())

	if cfg.FileStoragePath == "" {
		repository = storage.NewMapRepository()
	} else {
		repository, err = storage.NewFileRepository(serverCtx, cfg.FileStoragePath)
		if err != nil {
			panic(err)
		}
	}
	logger.Info("Created repository...", "type=", fmt.Sprintf("%T", repository))
	defer repository.Shutdown()

	service := services.NewShortenerService(generator, repository, shortURLLength)
	logger.Info("Created shorten service...")
	defer service.Shutdown()

	var dbCheckService srv.DBCheck

	if cfg.PostgresDSN != "" {
		db := sqlx.MustOpen("pgx", cfg.PostgresDSN)
		logger.Info("Created DB connection...")
		defer db.Close()

		dbCheckService = services.NewDBCheckService(db)
	} else {
		dbCheckService = &srv.NoOpDBCheck{}
	}

	logger.Info("Created DB check service...", "type=", fmt.Sprintf("%T", dbCheckService))
	defer dbCheckService.Shutdown()

	api := apis.NewShortenerAPI(service, dbCheckService, cfg.BaseURL)
	handler := api.Routes(logger)

	serverObj := server.NewServer(serverCtx, cfg.Addr, handler)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		s := <-signalChan
		logger.Info("Received os.Signal. Try graceful shutdown.", "signal=", s)

		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Error("Gracelful shutdown timeout. Force shutdown")
				os.Exit(1)
			}
		}()

		err := serverObj.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error("Error while shutdown", "err", err)
			os.Exit(1)
		}

		serverCancel()

		logger.Info("Server shutdowned")
	}()
	logger.Info("Starting listen and serve...", "addr=", serverObj.Addr)
	err = serverObj.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}
