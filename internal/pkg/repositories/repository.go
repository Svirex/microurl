package repositories

import (
	"context"
	"errors"

	"github.com/Svirex/microurl/internal/pkg/models"
)

var ErrNotFound = errors.New("not found record")
var ErrSomtheingWrong = errors.New("unknown error")

type URLRepository interface {
	Add(context.Context, *models.RepositoryAddRecord) (*models.RepositoryGetRecord, error)
	Get(context.Context, *models.RepositoryGetRecord) (*models.RepositoryGetResult, error)
	Shutdown() error
}
