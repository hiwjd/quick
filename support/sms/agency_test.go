package sms

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgencyWithMockSender(t *testing.T) {
	var mockSender = func(raw RawResult, err error) Sender {
		return func(number, content string) (RawResult, error) {
			return raw, err
		}
	}

	for _, tc := range []struct {
		raw RawResult
		err error
	}{
		{RawResult{"message": "ok"}, nil},
		{RawResult{"message": "gateway timeout"}, errors.New("gateway timeout")},
	} {
		agency, err := NewAgency(mockSender(tc.raw, tc.err))
		assert.Nil(t, err)
		result, err := agency.Send("15612341234", "hey there!")
		assert.Equal(t, tc.err, err)
		assert.Equal(t, tc.raw, result.Raw)
	}
}
