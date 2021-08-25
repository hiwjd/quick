package sqlex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUintListScanAndValue(t *testing.T) {
	var m UintList = []uint{1, 2}
	val, err := m.Value()
	assert.Nil(t, err)

	var m2 UintList
	err = m2.Scan(val)
	assert.Nil(t, err)
	assert.Equal(t, m, m2)

	bs := []byte(`[1,2]`)
	var m3 UintList
	err = m3.Scan(bs)
	assert.Nil(t, err)
	assert.Equal(t, m, m3)
}

func TestUintListEmptyShouldReturnEmpty(t *testing.T) {
	bs := []byte("")
	var m UintList
	err := m.Scan(bs)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, 0, len(m))
}
