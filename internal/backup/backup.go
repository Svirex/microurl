package backup

import "context"

type Record struct {
	UUID    string `json:"uuid"`
	ShortID string `json:"short_url"`
	URL     string `json:"original_url"`
}

type BackupReader interface {
	Read(context.Context) (*Record, error)
	Close() error
}

type BackupWriter interface {
	Write(context.Context, *Record) error
	Close() error
}
