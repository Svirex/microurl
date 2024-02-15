package util

type Generator interface {
	RandString(size uint) string
}
