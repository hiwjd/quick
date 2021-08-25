package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandLetters(t *testing.T) {
	s := RandLetters(4)
	assert.Equal(t, 4, len(s))

	s = RandLetters(8)
	assert.Equal(t, 8, len(s))
}

func TestRandNumbers(t *testing.T) {
	s := RandNumbers(4)
	assert.Equal(t, 4, len(s))

	s = RandNumbers(8)
	assert.Equal(t, 8, len(s))
}
