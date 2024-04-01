package models

type RepositoryAddRecord struct {
	ShortID string
	URL     string
	UID     string
}

type RepositoryGetRecord struct {
	ShortID string
}

type RepositoryGetResult struct {
	URL string
}

type UserURLRecord struct {
	URL     string
	ShortID string
}

func NewRepositoryAddRecord(shortID, url, uid string) *RepositoryAddRecord {
	return &RepositoryAddRecord{
		ShortID: shortID,
		URL:     url,
		UID:     uid,
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
