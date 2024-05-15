package config

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Addr            string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	PostgresDSN     string `env:"DATABASE_DSN"`
	MigrationsPath  string `env:"MIGRATIONS_PATH"`
	SecretKey       string `env:"SECRET_KEY"`
}

func ParseEnv() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse enviroment variables: %w", err)
	}
	return cfg, nil
}

func ParseFlags() (*Config, error) {
	cfg := &Config{}
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getwd: %w", err)
	}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&cfg.BaseURL, "b", "", "base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/short-url-db.json", "file for save records")
	flag.StringVar(&cfg.PostgresDSN, "d", "", "postgres DSN")
	flag.StringVar(&cfg.MigrationsPath, "m", path.Join(strings.Replace(currentDir, "cmd/shortener", "", 1), "migrations"), "path to db migrations")
	flag.StringVar(&cfg.SecretKey, "k", "fake_secret_key", "secret key for auth")
	flag.Parse()
	return cfg, nil
}

func Parse() (*Config, error) {
	envCfg, err := ParseEnv()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	flagConfig, err := ParseFlags()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	cfg := mergeConf(envCfg, flagConfig)
	prepareConfig(cfg)
	return cfg, nil
}

func prepareAddr(addr string) string {
	if strings.HasPrefix(addr, "http://") {
		return addr[7:]
	}
	return addr
}

func prepareBaseURL(baseURL, addr string) string {
	if baseURL != "" {
		if !strings.HasPrefix(baseURL, "http://") {
			return "http://" + baseURL
		}
		return baseURL
	}
	return "http://" + addr
}

func prepareConfig(cfg *Config) {
	cfg.Addr = prepareAddr(cfg.Addr)
	cfg.BaseURL = prepareBaseURL(cfg.BaseURL, cfg.Addr)
}

func mergeConf(envCfg *Config, flagConfig *Config) *Config {
	cfg := &Config{
		Addr:            envCfg.Addr,
		BaseURL:         envCfg.BaseURL,
		FileStoragePath: envCfg.FileStoragePath,
		PostgresDSN:     envCfg.PostgresDSN,
		MigrationsPath:  envCfg.MigrationsPath,
		SecretKey:       envCfg.SecretKey,
	}
	if cfg.Addr == "" {
		cfg.Addr = flagConfig.Addr
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = flagConfig.BaseURL
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = flagConfig.FileStoragePath
	}
	if cfg.PostgresDSN == "" {
		cfg.PostgresDSN = flagConfig.PostgresDSN
	}
	if cfg.MigrationsPath == "" {
		cfg.MigrationsPath = flagConfig.MigrationsPath
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = flagConfig.SecretKey
	}
	return cfg
}
