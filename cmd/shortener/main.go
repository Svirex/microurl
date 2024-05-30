package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Svirex/microurl/internal/adapters/api"
	"github.com/Svirex/microurl/internal/adapters/generator"
	repo "github.com/Svirex/microurl/internal/adapters/repository/postgres"
	"github.com/Svirex/microurl/internal/config"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/Svirex/microurl/internal/core/service"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const shortURLLength uint = 8

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      true,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	l, err := config.Build()
	if err != nil {
		log.Fatalf("couldn't init zap logger")
	}
	logger := ports.Logger(l.Sugar())
	defer logger.Sync()

	generator := generator.NewStringGenerator(time.Now().UnixNano())
	logger.Info("Created generator...")

	serverCtx, serverCancel := context.WithCancel(context.Background())

	var db *pgxpool.Pool
	if cfg.PostgresDSN != "" {
		logger.Info("Try create DB connection...")
		db, err = pgxpool.New(serverCtx, cfg.PostgresDSN)
		if err != nil {
			logger.Fatalln("DB connection error", "err", err)
		}
		logger.Info("DB connection success...")
		defer db.Close()

		migrationUp(db, logger, cfg.MigrationsPath)
	}

	repository := repo.NewPostgresRepository(db, logger)
	defer repository.Shutdown()
	logger.Infoln("Created repository...", "type=", fmt.Sprintf("%T", repository))

	shortenerService := service.NewShortenerService(generator, repository, shortURLLength, cfg.BaseURL)
	defer shortenerService.Shutdown()
	logger.Info("Created shorten service...")

	dbCheckService := service.NewDBCheck(db, cfg)
	logger.Info("Created DB check service...", "type=", fmt.Sprintf("%T", dbCheckService))

	deleterRepo := repo.NewDeleterRepository(db, logger)

	deleter, err := service.NewDeleter(deleterRepo, logger, 10)
	if err != nil {
		log.Fatalf("create deleter service: %#v", err)
	}
	deleter.Run()
	defer deleter.Shutdown()

	serviceAPI := api.NewAPI(shortenerService, dbCheckService, logger, deleter, cfg.SecretKey)
	handler := serviceAPI.Routes()

	serverObj := api.NewServer(serverCtx, cfg.Addr, handler)

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

func migrationUp(dbpool *pgxpool.Pool, logger ports.Logger, migrationsPath string) {
	db := stdlib.OpenDBFromPool(dbpool)
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Fatalf("create instance db for migrate: %v", "err", err)
	}
	migration, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath, "postgres", driver)
	if err != nil {
		logger.Fatalf("create migrate: %v", "err", err)
	}

	version, dirty, err := migration.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		logger.Fatalf("migration version error ", "err=", err)
	}
	if dirty {
		err = migration.Force(int(version))
		if err != nil {
			logger.Fatalf("migration force error ", "err=", err)
		}
	}
	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatalf("migration up error ", "err=", err)
	}
}
