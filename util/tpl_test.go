package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTpl(t *testing.T) {
	s := Tpl("Hello, {{name}}", TplParam{"name": "simon"})
	assert.Equal(t, "Hello, simon", s)
}

func BenchmarkTpl(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := Tpl("Hello, {{name}}", TplParam{"name": "simon"})
		if s != "Hello, simon" {
			b.FailNow()
		}
	}
}
