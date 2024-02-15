package services

import "github.com/Svirex/microurl/internal/pkg/models"

type Shortener interface {
	Add(*models.ServiceAddRecord) (*models.ServiceAddResult, error)
	Get(*models.ServiceGetRecord) (*models.ServiceGetResult, error)
}
