package config

import (
	"flag"
	"log"
	"strings"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Addr    string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func ParseEnv() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}

func ParseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&cfg.BaseURL, "b", "", "base URL")
	flag.Parse()
	return cfg
}

func Parse() *Config {
	envCfg := ParseEnv()
	flagConfig := ParseFlags()
	cfg := mergeConf(envCfg, flagConfig)
	prepareConfig(cfg)
	return cfg
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
		Addr:    envCfg.Addr,
		BaseURL: envCfg.BaseURL,
	}
	if cfg.Addr == "" {
		cfg.Addr = flagConfig.Addr
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = flagConfig.BaseURL
	}
	return cfg
}
