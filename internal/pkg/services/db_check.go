package services

import "context"

type DBCheck interface {
	Ping(context.Context) error
	Shutdown() error
}
