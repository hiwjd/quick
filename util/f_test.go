package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestF322d(t *testing.T) {
	v := F322d(3.1999998)
	assert.EqualValues(t, 3.20, v)
}
