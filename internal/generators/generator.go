package generators

type Generator interface {
	RandString(size int) string
}
