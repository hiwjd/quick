package quick

import (
	"net/http"

	"github.com/go-sql-driver/mysql"
	"github.com/hiwjd/quick/util"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type customHTTPErrorHandler struct {
	e    *echo.Echo
	logf Logf
}

// NewCustomHTTPErrorHandler 构造 echo.HTTPErrorHandler
func NewCustomHTTPErrorHandler(e *echo.Echo, logf Logf) echo.HTTPErrorHandler {
	cheh := &customHTTPErrorHandler{e: e, logf: logf}
	return cheh.Handle
}

// Handle 实现 echo.HTTPErrorHandler
func (cheh *customHTTPErrorHandler) Handle(err error, c echo.Context) {
	code := http.StatusInternalServerError
	body := map[string]string{"message": http.StatusText(code)}
	caller := ""

	switch t := err.(type) {
	case *mysql.MySQLError:
		switch t.Number {
		case 1062:
			// 键冲突
			code = http.StatusBadRequest
			body["message"] = t.Message
			break
		}
		break
	case ErrorArray:
		code = http.StatusBadRequest
		body = map[string]string{"message": t.Error()}
		break
	case *FineErr:
		code = t.Code
		body = map[string]string{"message": t.Message}
		caller = t.Caller
		break
	case BizErr:
		switch t.Type {
		case BadArgs:
			code = http.StatusBadRequest
		case SrvErr:
			code = http.StatusInternalServerError
		default:
			code = http.StatusInternalServerError
		}
		body = map[string]string{"message": t.Error()}
		caller = util.WrapCaller(t.Caller())
		if t.cause != nil {
			cheh.logf("%#v", t.cause)
		}
	default:
		switch err {
		case gorm.ErrRecordNotFound:
			code = http.StatusBadRequest
			body = map[string]string{"message": http.StatusText(code)}
			_, _, caller = util.Caller(3)
			break
		default:
			cheh.e.DefaultHTTPErrorHandler(err, c)
			return
		}
	}

	cheh.logf("[ERROR] %s: %d %s\n", caller, code, body["message"])

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, body)
		}
		if err != nil {
			cheh.logf("[ERROR] %s\n", err.Error())
		}
	}
}
