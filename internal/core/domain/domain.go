package domain

type ShortID string
type ShortURL string
type URL string
type UID string

type Record struct {
	UID UID
	URL URL
}

type URLData struct {
	URL      URL      `json:"original_url"`
	ShortID  ShortID  `json:"-"`
	ShortURL ShortURL `json:"short_url"`
}

type BatchRecord struct {
	CorrID   string   `json:"correlation_id"`
	URL      URL      `json:"original_url"`
	ShortID  ShortID  `json:"-"`
	ShortURL ShortURL `json:"short_url"`
}
