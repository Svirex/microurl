package models

type ServiceAddRecord struct {
	URL string
	UID string
}

type ServiceGetRecord struct {
	ShortID string
}

type ServiceAddResult struct {
	ShortID string
}

type ServiceGetResult struct {
	URL string
}

func NewServiceAddRecord(url string, uid string) *ServiceAddRecord {
	return &ServiceAddRecord{
		URL: url,
		UID: uid,
	}
}

func NewServiceGetRecord(shortID string) *ServiceGetRecord {
	return &ServiceGetRecord{
		ShortID: shortID,
	}
}

func NewServiceGetResult(url string) *ServiceGetResult {
	return &ServiceGetResult{
		URL: url,
	}
}

func NewServiceAddResult(shortID string) *ServiceAddResult {
	return &ServiceAddResult{
		ShortID: shortID,
	}
}

type BatchRequestRecord struct {
	CorrID string `json:"correlation_id"`
	URL    string `json:"original_url"`
}

type BatchRequest struct {
	Records []BatchRequestRecord
}

type BatchResponseRecord struct {
	CorrID   string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

type BatchServiceRecord struct {
	CorrID   string `json:"correlation_id"`
	URL      string `json:"original_url"`
	ShortURL string `json:"short_url"`
}

type BatchService struct {
	Records []BatchServiceRecord
}

type BatchResponse struct {
	Records []BatchResponseRecord
}
