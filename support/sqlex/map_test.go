package sqlex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapScanAndValue(t *testing.T) {
	var m Map = map[string]interface{}{"产地": "希腊", "净含量": "800克", "存储条件": "冷藏"}
	val, err := m.Value()
	assert.Nil(t, err)

	var m2 Map
	err = m2.Scan(val)
	assert.Nil(t, err)
	assert.Equal(t, m, m2)

	bs := []byte(`{"产地": "希腊", "净含量": "800克", "存储条件": "冷藏"}`)
	var m3 Map
	err = m3.Scan(bs)
	assert.Nil(t, err)
	assert.Equal(t, m, m3)
}

func TestMapEmptyShouldReturnEmptyMap(t *testing.T) {
	bs := []byte("")
	var m Map
	err := m.Scan(bs)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, 0, len(m))
}
