package models

type RepositoryAddRecord struct {
	ShortID string
	URL     string
}

type RepositoryGetRecord struct {
	ShortID string
}

type RepositoryGetResult struct {
	URL string
}

func NewRepositoryAddRecord(shortID, url string) *RepositoryAddRecord {
	return &RepositoryAddRecord{
		ShortID: shortID,
		URL:     url,
	}
}

func NewRepositoryGetRecord(shortdID string) *RepositoryGetRecord {
	return &RepositoryGetRecord{
		ShortID: shortdID,
	}
}

func NewRepositoryGetResult(url string) *RepositoryGetResult {
	return &RepositoryGetResult{
		URL: url,
	}
}
