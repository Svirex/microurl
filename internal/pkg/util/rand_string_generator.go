package util

type Generator interface {
	RandString(size int) string
}
