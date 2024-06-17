package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
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

	serverObj := api.NewServer(serverCtx, handler)

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
	listener, err := createListner(cfg.EnableHTTPS, cfg.Addr)
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

func createListner(enableHTTPS bool, addr string) (net.Listener, error) {
	if !enableHTTPS {
		return net.Listen("tcp", addr)
	}
	return createTLSListener(addr)
}

func createTLSListener(addr string) (net.Listener, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Svirex"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("unable generate rsa key: %w", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable create x509 cert: %w", err)
	}

	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	certPair, err := tls.X509KeyPair(certPEM.Bytes(), privateKeyPEM.Bytes())
	if err != nil {
		return nil, fmt.Errorf("unable create x509 pair: %w", err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{certPair}}
	listener, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable create tls listener: %w", err)
	}

	return listener, nil
}
