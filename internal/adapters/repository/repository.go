package repository

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Svirex/microurl/internal/adapters/filebackup"
	"github.com/Svirex/microurl/internal/adapters/repository/file"
	"github.com/Svirex/microurl/internal/adapters/repository/inmemory"
	repo "github.com/Svirex/microurl/internal/adapters/repository/postgres"
	"github.com/Svirex/microurl/internal/config"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewRepository(ctx context.Context, cfg *config.Config, db *pgxpool.Pool, logger ports.Logger) (ports.ShortenerRepository, error) {
	var repository ports.ShortenerRepository
	if cfg.PostgresDSN != "" {
		repository = repo.NewPostgresRepository(db, logger)
		migrationUp(db, logger, cfg.MigrationsPath)
	} else if cfg.FileStoragePath != "" {
		f, err := os.OpenFile(cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			return nil, fmt.Errorf("new repository, open file: %w", err)
		}
		m := inmemory.NewShortenerRepository()
		r := filebackup.NewFileBackupReader(f)
		err = r.Restore(ctx, m)
		if err != nil {
			return nil, fmt.Errorf("new repository, restore file backup: %w", err)
		}
		f.Close()
		f, err = os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("new repository, open file after restore: %w", err)
		}

		w := filebackup.NewFileBackupWriter(f)
		repository = file.NewShortenerRepository(m, w)

	} else {
		repository = inmemory.NewShortenerRepository()
	}
	return repository, nil
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
