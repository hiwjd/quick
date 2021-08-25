package sqlex

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringListScanAndValue(t *testing.T) {
	var sList StringList = []string{"aaa", "bbb"}
	val, err := sList.Value()
	assert.Nil(t, err)
	assert.Equal(t, "string", reflect.TypeOf(val).Name())

	var sList2 StringList
	err = sList2.Scan(val)
	assert.Nil(t, err)
	assert.Equal(t, sList, sList2)

	bs := []byte(`aaa,bbb`)
	var sList3 StringList
	err = sList3.Scan(bs)
	assert.Nil(t, err)
	assert.Equal(t, sList, sList2)
}

type req struct {
	StrList StringList `json:"strList"`
}

func TestStringListJsonMarshalAndUnmarshal(t *testing.T) {
	bs := []byte(`{"strList":["aaa","bbb"]}`)

	var r req
	err := json.Unmarshal(bs, &r)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r.StrList))
	assert.Equal(t, "aaa", r.StrList[0])
	assert.Equal(t, "bbb", r.StrList[1])
}

func TestStringListEmptyStringShouldReturnEmptySlice(t *testing.T) {
	bs := []byte("")
	var sList StringList
	err := sList.Scan(bs)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(sList))
}
