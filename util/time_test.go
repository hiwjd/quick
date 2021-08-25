package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBeginningOfMonth(t *testing.T) {
	dt, _ := time.Parse("2006-01-02", "2020-08-10")
	bm := BeginningOfMonth(dt)
	assert.Equal(t, 2020, bm.Year())
	assert.Equal(t, time.Month(8), bm.Month())
	assert.Equal(t, 1, bm.Day())
	assert.Equal(t, 0, bm.Hour())
	assert.Equal(t, 0, bm.Minute())
	assert.Equal(t, 0, bm.Second())
}

func TestEndOfMonth(t *testing.T) {
	dt, _ := time.Parse("2006-01-02", "2020-08-10")
	bm := EndOfMonth(dt)
	assert.Equal(t, 2020, bm.Year())
	assert.Equal(t, time.Month(8), bm.Month())
	assert.Equal(t, 31, bm.Day())
	assert.Equal(t, 23, bm.Hour())
	assert.Equal(t, 59, bm.Minute())
	assert.Equal(t, 59, bm.Second())
}
