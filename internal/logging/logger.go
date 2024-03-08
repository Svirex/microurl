package logging

import (
	"github.com/Svirex/microurl/internal/pkg/logging"
	"go.uber.org/zap"
)

var _ logging.Logger = (*DefaultLogger)(nil)

type DefaultLogger struct {
	Logger *zap.SugaredLogger
}

func (logger *DefaultLogger) Info(params ...any) {
	logger.Logger.Infoln(params...)
}

func NewDefaultLogger() (logging.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return &DefaultLogger{
		Logger: logger.Sugar(),
	}, nil
}
