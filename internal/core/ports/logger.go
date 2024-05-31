package ports

import (
	"go.uber.org/zap"
)

// Logger - тип для логгера
type Logger = *zap.SugaredLogger
