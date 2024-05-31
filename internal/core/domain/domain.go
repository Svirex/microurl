// Пакет domain описывает основные сущности, используемые в проекте.
package domain

// ShortID - тип для короткого идентификатора.
type ShortID string

// ShortURL - тип для сокращенного урла.
type ShortURL string

// URL - тип для урла.
type URL string

// UID - тип для uid пользователя.
type UID string

// Record определяет тип для записи к БД.
type Record struct {
	UID UID
	URL URL
}

// URLData - тип записи реального URL и сокращенного URL.
type URLData struct {
	URL      URL     `json:"original_url"`
	ShortID  ShortID `json:"-"`
	ShortURL URL     `json:"short_url"`
}

// BatchRecord - тип записи при добавления записей батчей.
type BatchRecord struct {
	CorrID   string  `json:"correlation_id"`
	URL      URL     `json:"original_url"`
	ShortID  ShortID `json:"-"`
	ShortURL URL     `json:"short_url"`
}

// BackupRecord - тип для хранения данных в файле.
type BackupRecord struct {
	UUID    string  `json:"uuid"`
	ShortID ShortID `json:"short_url"`
	URL     URL     `json:"original_url"`
	UID     UID     `json:"uid,omitempty"`
}

// DeleteData - данные для пометки URL как удаленного.
type DeleteData struct {
	UID     string
	ShortID string
}
