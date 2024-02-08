package main

import (
	"flag"
	"strings"

	"github.com/Svirex/microurl/internal/pkg/config"
)

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

func prepareConfig(cfg *config.Config) {
	cfg.Addr = prepareAddr(cfg.Addr)
	cfg.BaseURL = prepareBaseURL(cfg.BaseURL, cfg.Addr)
}

func parseFlagsIntoConfig() *config.Config {
	cfg := &config.Config{}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&cfg.BaseURL, "b", "", "base URL")
	flag.Parse()
	prepareConfig(cfg)
	return cfg
}
