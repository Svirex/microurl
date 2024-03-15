package services

import (
	"context"

	"github.com/Svirex/microurl/internal/pkg/models"
)

type Shortener interface {
	Add(context.Context, *models.ServiceAddRecord) (*models.ServiceAddResult, error)
	Get(context.Context, *models.ServiceGetRecord) (*models.ServiceGetResult, error)
	Batch(context.Context, *models.BatchRequest) (*models.BatchResponse, error)
	Shutdown() error
}
