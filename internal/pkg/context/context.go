package context

import (
	"github.com/Svirex/microurl/internal/pkg/config"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/pkg/util"
)

type AppContext struct {
	Config     *config.Config
	Repository repositories.Repository
	Generator  util.Generator
}
