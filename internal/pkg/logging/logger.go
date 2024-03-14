package logging

type Logger interface {
	Info(params ...any)
	Error(params ...any)
	Shutdown() error
}
