package quick

import (
	"fmt"
	"net/http"

	"github.com/hiwjd/quick/util"
)

// FineErr 是一个可以对用户友好点的错误
type FineErr struct {
	err     error
	Code    int
	Message string
	Caller  string
}

// NewFineErr 构造FineErr
func NewFineErr(code int, message string) *FineErr {
	file, line, fn := util.Caller(3)
	caller := fmt.Sprintf("%s:%d %s", file, line, fn)
	return &FineErr{Code: code, Message: message, Caller: caller}
}

func (fe *FineErr) Error() string {
	if fe.err == nil {
		return fe.Message
	}
	return fe.err.Error()
}

// SetError 设置err
func (fe *FineErr) SetError(err error) *FineErr {
	fe.err = err
	return fe
}

func TraceError(err error) *FineErr {
	return NewFineErr(http.StatusInternalServerError, err.Error()).SetError(err)
}

type ErrType int

const (
	BadArgs ErrType = iota
	SrvErr
)

type BizErr struct {
	Type  ErrType
	cause error
	code  string
	file  string
	line  int
	fn    string
}

func NewBadArgs(code string, cause error) BizErr {
	return newBizErr(BadArgs, code, cause)
}

func NewSrvErr(code string, cause error) BizErr {
	return newBizErr(SrvErr, code, cause)
}

func newBizErr(typ ErrType, code string, cause error) BizErr {
	file, line, fn := util.Caller(4)
	return BizErr{
		Type:  typ,
		cause: cause,
		code:  code,
		file:  file,
		line:  line,
		fn:    fn,
	}
}

func (be BizErr) Error() string {
	return be.code
}

func (be BizErr) Unwrap() error {
	return be.cause
}

func (be BizErr) Is(err error) bool {
	switch err.(type) {
	case BizErr:
		return true
	}
	return false
}

func (be BizErr) Caller() (file string, line int, fn string) {
	return be.file, be.line, be.fn
}
