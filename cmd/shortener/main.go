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
	"github.com/Svirex/microurl/internal/adapters/repository"
	repo "github.com/Svirex/microurl/internal/adapters/repository/postgres"
	"github.com/Svirex/microurl/internal/config"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/Svirex/microurl/internal/core/service"
	"github.com/Svirex/microurl/internal/server"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const shortURLLength uint = 8

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func showMetadata() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)
}

func main() {
	showMetadata()
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cfg)
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
		log.Panicln("couldn't init zap logger")
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
			logger.Panicln("DB connection error", "err", err)
		}
		logger.Info("DB connection success...")

		closeDB := func() {
			logger.Debug("start close db")
			db.Close()
			logger.Debug("end close db")
		}
		defer closeDB()
	}

	repository, err := repository.NewRepository(serverCtx, cfg, db, logger)
	if err != nil {
		logger.Panicf("create repository err: %w\n", err)
	}
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
		logger.Panicf("create deleter service: %#v", err)
	}
	deleter.Run()
	defer deleter.Shutdown()

	serviceAPI := api.NewAPI(shortenerService, dbCheckService, logger, deleter, cfg.SecretKey)
	handler := serviceAPI.Routes()

	serverObj := server.NewServer(serverCtx, handler)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		s := <-signalChan
		logger.Info("Received os.Signal. Try graceful shutdown.", "signal=", s)

		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		logger.Debug("start shutdown server")

		err := serverObj.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error("Error while shutdown", "err", err)
		}

		logger.Debug("start serverCancel")

		serverCancel()

		logger.Info("Server shutdowned")
	}()
	listener, err := server.CreateListener(cfg.EnableHTTPS, cfg.Addr)
	if err != nil {
		logger.Panicf("create listener: %#v", err)
	}
	logger.Info("Starting server on addr ", listener.Addr())
	err = serverObj.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf("Serve: %v", err)
	}

	<-serverCtx.Done()
}
