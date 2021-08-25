package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleAESEncDec(t *testing.T) {
	key := []byte("1234567890123456")
	simpleAES, err := NewSimpleAES(key)
	assert.Nil(t, err)

	data := []byte("some plain text need enc")
	cipherData, err := simpleAES.Enc(data)
	assert.Nil(t, err)

	plainData, err := simpleAES.Dec(cipherData)
	assert.Nil(t, err)

	assert.Equal(t, data, plainData)
}

func BenchmarkSimpleAES(b *testing.B) {
	key := []byte("1234567890123456")
	simpleAES, _ := NewSimpleAES(key)
	b.ResetTimer()

	data := []byte("some plain text need enc")
	for i := 0; i < b.N; i++ {
		cd, _ := simpleAES.Enc(data)
		d, _ := simpleAES.Dec(cd)
		if string(d) != string(data) {
			b.FailNow()
		}
	}
}
