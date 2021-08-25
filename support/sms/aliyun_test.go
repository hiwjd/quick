package sms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAliyunSender(t *testing.T) {
	t.Skip("阿里云短信跳过测试")
	sender, err := NewAliyunSender("cn-hangzhou", "", "", "")
	assert.Nil(t, err)

	rsp, err := sender.Send("15912341234", "SMS_142100306", `{"code":"1122"}`)
	assert.Nil(t, err)

	assert.Equal(t, "OK", rsp.Code)
}
