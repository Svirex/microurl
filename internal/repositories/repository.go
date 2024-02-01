package repositories

type Repository interface {
	Add(shortId, url string) error
	Get(shortId string) (string, error)
}
