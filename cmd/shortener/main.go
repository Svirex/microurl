package main

import (
	"github.com/Svirex/microurl/internal/pkg/app"
	"github.com/Svirex/microurl/internal/pkg/config"
)

func main() {
	cfg := config.Parse()
	err := app.Run(cfg)
	if err != nil {
		panic(err)
	}
}
