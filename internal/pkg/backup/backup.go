package backup

type Record struct {
	UUID    string `json:"uuid"`
	ShortID string `json:"short_url"`
	URL     string `json:"original_url"`
}

type Backup interface {
	Write(record *Record) error
	Read() (*Record, error)
}
