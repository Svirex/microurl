package services

import (
	"context"
	"errors"
)

var ErrPingFailed = errors.New("ping to db failed")

type DBCheck interface {
	Ping(context.Context) error
	Shutdown() error
}

type NoOpDBCheck struct{}

var _ DBCheck = (*NoOpDBCheck)(nil)

func (n *NoOpDBCheck) Ping(context.Context) error {
	return ErrPingFailed
}

func (n *NoOpDBCheck) Shutdown() error {
	return nil
}
