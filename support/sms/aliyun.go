package sms

import (
	"errors"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

// AliyunSender 是阿里云短信实现
type AliyunSender struct {
	signName string
	client   *dysmsapi.Client
}

// NewAliyunSender 构造AliyunSender
func NewAliyunSender(regionID, accessKeyID, accessSecret, signName string) (*AliyunSender, error) {
	client, err := dysmsapi.NewClientWithAccessKey(regionID, accessKeyID, accessSecret)
	if err != nil {
		return nil, err
	}

	return &AliyunSender{
		signName: signName,
		client:   client,
	}, nil
}

// Send 发送短信
func (s *AliyunSender) Send(number, tplCode, tplParam string) (*dysmsapi.SendSmsResponse, error) {
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.SignName = s.signName
	request.PhoneNumbers = number
	request.TemplateCode = tplCode
	request.TemplateParam = tplParam

	return s.client.SendSms(request)
}

// AsSender 包装成Sender
func (s *AliyunSender) AsSender() Sender {
	return func(number, content string) (RawResult, error) {
		// 在阿里云短信内，content这样组成：模板ID:{"code":"1234"}
		parts := strings.SplitN(content, ":", 2)
		if len(parts) != 2 {
			return RawResult{"error": "解析阿里云短信所需参数异常"}, errors.New("解析阿里云短信所需参数异常")
		}

		rsp, err := s.Send(number, parts[0], parts[1])
		result := RawResult{
			"BizId":     rsp.BizId,
			"Code":      rsp.Code,
			"Message":   rsp.Message,
			"RequestId": rsp.RequestId,
		}

		return result, err
	}
}
