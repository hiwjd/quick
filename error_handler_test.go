package quick

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type tcase struct {
	err      error
	wantBody []byte
	wantCode int
}

func TestCustomHttpErrorHandler(t *testing.T) {
	e := echo.New()
	handle := NewCustomHTTPErrorHandler(e, func(format string, args ...interface{}) {})

	cases := []tcase{
		{err: errors.New("normal"), wantBody: []byte(`{"message":"Internal Server Error"}`), wantCode: 500},
		{err: &mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'x' for key y"}, wantBody: []byte(`{"message":"Duplicate entry 'x' for key y"}`), wantCode: 400},
		{err: echo.NewHTTPError(404, "not found"), wantBody: []byte(`{"message":"not found"}`), wantCode: 404},
		{err: ErrorArray{ErrorData{Field: "name", Title: "姓名", Errors: []string{"不能为空"}}}, wantBody: []byte(`{"message":"姓名: 不能为空"}`), wantCode: 400},
		{err: NewFineErr(401, "unauth"), wantBody: []byte(`{"message":"unauth"}`), wantCode: 401},
		{err: gorm.ErrRecordNotFound, wantBody: []byte(`{"message":"Not Found"}`), wantCode: 404},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handle(tc.err, c)

		assert.Equal(t, tc.wantCode, rec.Code)
		body, _ := ioutil.ReadAll(rec.Result().Body)
		assert.Equal(t, tc.wantBody, body[:len(body)-1]) // 读取到响应体中会多一个 0x0a，是json序列化时多加的'\n'，见 encoding/json/stream.go#L211 e.WriteByte('\n')
	}

}
