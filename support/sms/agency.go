package sms

// RawResult 是发送器未经加工的发送结果
type RawResult map[string]string

// Sender 表示一个短信发送器
type Sender func(number, content string) (RawResult, error)

// Agency 表示一个短信中介
type Agency interface {

	// Send 发送短信
	// 无论是否发送成功，SMSResult都有值，成功是error是nil，否则有值
	Send(number, content string) (*Result, error)
}

// Result 是发送结果
type Result struct {
	Raw RawResult
}

// NewAgency 构造Sender
func NewAgency(sender Sender) (Agency, error) {
	return &simpleSender{
		sender: sender,
	}, nil
}

type simpleSender struct {
	sender Sender
}

func (s *simpleSender) Send(number, content string) (*Result, error) {
	m, err := s.sender(number, content)

	return &Result{
		Raw: m,
	}, err
}
