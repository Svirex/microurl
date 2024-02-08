package main

import (
	"github.com/Svirex/microurl/internal/pkg/app"
	"github.com/Svirex/microurl/internal/pkg/config"
)

func main() {
	cfg := config.Parse()
	app.Run(cfg.Addr, cfg.BaseURL)
}
