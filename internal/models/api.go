package models

type InputJSON struct {
	URL string `json:"url"`
}

type ResultJSON struct {
	ShortURL string `json:"result"`
}

type UserURL struct {
	URL      string `json:"original_url"`
	ShortURL string `json:"short_url"`
}
