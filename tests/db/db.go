// Пакет испоьзуетсяд ля запуска интеграционных тестов.
package db

import (
	"context"
	"errors"
	"log"

	"github.com/Svirex/microurl/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var dbpool *pgxpool.Pool
var logger *zap.SugaredLogger

var migration *migrate.Migrate

// GetPool - получить пул соединений к БД.
func GetPool() *pgxpool.Pool {
	if dbpool == nil {
		log.Fatalf("db not init")
	}
	return dbpool
}

// GetLogger - получить логгер.
func GetLogger() *zap.SugaredLogger {
	if logger == nil {
		log.Fatalf("logger not init")
	}
	return logger
}

func initLogger() {
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
	logger = l.Sugar()
}

func initMigration(path string) {
	db := stdlib.OpenDBFromPool(dbpool)
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Fatalf("create instance db for migrate: %v", "err", err)
	}
	migration, err = migrate.NewWithDatabaseInstance(
		"file://"+path, "postgres", driver)
	if err != nil {
		logger.Fatalf("create migrate: %v", "err", err)
	}
}

// Init - создать сущности.
func Init() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	if logger == nil {
		initLogger()
	}
	if dbpool == nil {
		сonnect(cfg.PostgresDSN)
	}
	if migration == nil {
		initMigration(cfg.MigrationsPath)
	}
}

func сonnect(dsn string) {
	var err error
	dbpool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("create new pgxpool: %s, err: %v", dsn, err)
	}
	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatalf("db ping error: %v", err)
	}
	log.Println("DB Connected")
}

// Сlose - закрыть соединения.
func Сlose() {
	dbpool.Close()
}

// MigrateUp - накатить миграции.
func MigrateUp() {
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

// MigrateDown - откатить миграции.
func MigrateDown() {
	err := migration.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Info("couldn't down migration ", "err=", err)
	}
}

// Truncate - очистить таблицы в БД.
func Truncate() error {
	_, err := dbpool.Exec(context.Background(), "TRUNCATE TABLE users, records RESTART IDENTITY;")
	if err != nil {
		logger.Error("couldn't truncate tables ", err)
		return err
	}
	return nil
}
