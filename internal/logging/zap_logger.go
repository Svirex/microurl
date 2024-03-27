package logging

import (
	"go.uber.org/zap"
)

var _ Logger = (*DefaultLogger)(nil)

type DefaultLogger struct {
	logger *zap.SugaredLogger
}

func NewDefaultLogger() (Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return &DefaultLogger{
		logger: logger.Sugar(),
	}, nil
}

func (logger *DefaultLogger) Shutdown() error {
	return logger.logger.Sync()
}

func (logger *DefaultLogger) Info(params ...any) {
	logger.logger.Infoln(params...)
}

func (logger *DefaultLogger) Error(params ...any) {
	logger.logger.Errorln(params...)
}