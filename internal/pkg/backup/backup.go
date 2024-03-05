package backup

type Record struct {
	UUID    string `json:"uuid"`
	ShortID string `json:"short_url"`
	URL     string `json:"original_url"`
}

type BackupReader interface {
	Read() (*Record, error)
	Close() error
}

type BackupWriter interface {
	Write(*Record) error
	Close() error
}
