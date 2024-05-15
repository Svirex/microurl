package generators

import (
	"math/rand"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var _ Generator = (*SimpleGenerator)(nil)

type SimpleGenerator struct {
	Rand *rand.Rand
}

func (g *SimpleGenerator) RandString(size uint) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[g.Rand.Intn(len(letters))]
	}
	return string(b)
}

func NewSimpleGenerator(seed int64) Generator {
	return &SimpleGenerator{
		Rand: rand.New(rand.NewSource(seed)),
	}
}
