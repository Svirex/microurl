package generator

import (
	"context"
	"math/rand"

	"github.com/Svirex/microurl/internal/core/ports"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type StringGenerator struct {
	rand *rand.Rand
}

var _ ports.StringGenerator = (*StringGenerator)(nil)

func (g *StringGenerator) Generate(ctx context.Context, size uint) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[g.rand.Intn(len(letters))]
	}
	return string(b)
}

func NewStringGenerator(seed int64) *StringGenerator {
	return &StringGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}
