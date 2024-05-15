package generators

type Generator interface {
	RandString(size uint) string
}
