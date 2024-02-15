package repositories

import (
	"errors"

	"github.com/Svirex/microurl/internal/pkg/models"
)

var ErrNotFound = errors.New("not found record")
var ErrSomtheingWrong = errors.New("unknown error")

type Repository interface {
	Add(*models.RepositoryAddRecord) error
	Get(*models.RepositoryGetRecord) (*models.RepositoryGetResult, error)
}
