package main

import (
	"fmt"

	"github.com/Svirex/microurl/internal/pkg/app"
	"github.com/Svirex/microurl/internal/pkg/config"
)

func main() {
	cfg := config.Parse()
	fmt.Println(cfg.Addr, cfg.BaseURL)
	app.Run(cfg.Addr, cfg.BaseURL)
}
