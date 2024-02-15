package models

type ServiceAddRecord struct {
	URL string
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

func NewServiceAddRecord(url string) *ServiceAddRecord {
	return &ServiceAddRecord{
		URL: url,
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
