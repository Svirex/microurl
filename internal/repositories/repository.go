package repositories

type Repository interface {
	Add(shortID, url string) error
	Get(shortID string) (*string, error)
}
