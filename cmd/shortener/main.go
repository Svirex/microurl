package main

import (
	"fmt"

	"github.com/Svirex/microurl/internal/pkg/app"
)

func main() {
	cfg := parseFlagsIntoConfig()
	fmt.Println(cfg.Addr, cfg.BaseURL)
	app.Run(cfg.Addr, cfg.BaseURL)
}
