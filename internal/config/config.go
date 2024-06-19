package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/caarlos0/env/v10"
)

// Config - конфиг приложения
type Config struct {
	// Addr - адрес, по которому будет запущенно приложение
	Addr string `env:"SERVER_ADDRESS" json:"server_address"`
	// BaseURL - адрес, который будет использоваться для сокращенной ссылки
	BaseURL string `env:"BASE_URL" json:"base_url"`
	// FileStoragePath - путь к файлу для сохранения и загрузки записей
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	// PostgresDSN - параметры для подключения к БД Postgres
	PostgresDSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// MigrationsPath - путь до директории с файлами миграций БД
	MigrationsPath string `env:"MIGRATIONS_PATH"`
	// SecretKey - секретный ключ для создания JWT токена
	SecretKey string `env:"SECRET_KEY"`
	// EnableHTTPS - включить https
	EnableHTTPS bool `env:"ENABLE_HTTPS" json:"enable_https"`
	// ConfigPath - путь до файла с json-конфигом
	ConfigPath string `env:"CONFIG"`
}

// ParseEnv - парсим переменные окружения
func ParseEnv() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse enviroment variables: %w", err)
	}
	return cfg, nil
}

// ParseFlags - парсим флаги командной строки
func ParseFlags() (*Config, error) {
	cfg := &Config{}
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getwd: %w", err)
	}
	flag.StringVar(&cfg.Addr, "a", "", "<host>:<port>")
	flag.StringVar(&cfg.BaseURL, "b", "", "base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file for save records")
	flag.StringVar(&cfg.PostgresDSN, "d", "", "postgres DSN")
	flag.StringVar(&cfg.MigrationsPath, "m", path.Join(strings.Replace(currentDir, "cmd/shortener", "", 1), "migrations"), "path to db migrations")
	flag.StringVar(&cfg.SecretKey, "k", "fake_secret_key", "secret key for auth")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable https")
	flag.StringVar(&cfg.ConfigPath, "config", "", "path to json config")
	flag.StringVar(&cfg.ConfigPath, "c", "", "path to json config")
	flag.Parse()
	return cfg, nil
}

// Parse - парсим конфиг
func Parse() (*Config, error) {
	envCfg, err := ParseEnv()
	if err != nil {
		return nil, fmt.Errorf("parse env error: %w", err)
	}
	flagConfig, err := ParseFlags()
	if err != nil {
		return nil, fmt.Errorf("parse flag error: %w", err)
	}
	cfg := mergeConf(envCfg, flagConfig)
	if cfg.ConfigPath != "" {
		fileCfg, err := loadJSONConfig(cfg.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("file config read error: %w", err)
		}
		cfg = mergeCommonsFields(cfg, fileCfg)
	}
	setDefaults(cfg)
	prepareConfig(cfg)
	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:8080"
	}
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

// mergeCommonsFields - делает left join двух конфигов. Используется для общих полей настроек из переменных окружения,
// флагов и из файла
func mergeCommonsFields(first, second *Config) *Config {
	cfg := &Config{
		Addr:            second.Addr,
		BaseURL:         second.BaseURL,
		FileStoragePath: second.FileStoragePath,
		PostgresDSN:     second.PostgresDSN,
		EnableHTTPS:     second.EnableHTTPS,
	}
	if first.Addr != "" {
		cfg.Addr = first.Addr
	}
	if first.BaseURL != "" {
		cfg.BaseURL = first.BaseURL
	}
	if first.FileStoragePath != "" {
		cfg.FileStoragePath = first.FileStoragePath
	}
	if first.PostgresDSN != "" {
		cfg.PostgresDSN = first.PostgresDSN
	}
	if first.EnableHTTPS {
		cfg.EnableHTTPS = first.EnableHTTPS
	}
	return cfg
}

func mergeConf(envCfg *Config, flagConfig *Config) *Config {
	cfg := mergeCommonsFields(envCfg, flagConfig)
	cfg.MigrationsPath = flagConfig.MigrationsPath
	cfg.SecretKey = flagConfig.SecretKey
	cfg.ConfigPath = flagConfig.ConfigPath
	if envCfg.MigrationsPath != "" {
		cfg.MigrationsPath = envCfg.MigrationsPath
	}
	if envCfg.SecretKey != "" {
		cfg.SecretKey = envCfg.SecretKey
	}
	if envCfg.ConfigPath != "" {
		cfg.ConfigPath = envCfg.ConfigPath
	}
	return cfg
}

func loadJSONConfig(fileName string) (*Config, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}
	cfg := &Config{}
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error^ %w", err)
	}
	return cfg, nil
}
