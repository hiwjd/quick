package quick

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// CustomValidator 用于参数校验
type customValidator struct{}

// NewCustomValidator 构造echo.Validator实例
func NewCustomValidator() echo.Validator {
	return &customValidator{}
}

// Validate 实现echo.Validator
func (cv *customValidator) Validate(i interface{}) error {
	if err := Check(i); err != nil {
		// if err := validator.Validate(i); err != nil {
		// err 是 validator.ErrorMap 可以做更好的错误信息包装
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
