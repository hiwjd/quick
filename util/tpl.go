package util

import (
	"github.com/valyala/fasttemplate"
)

// TplParam 便于书写map
type TplParam map[string]interface{}

// Tpl 格式化文本
// "Hello, {{name}}" {"name":"simon"} => Hello, simon
func Tpl(template string, param TplParam) string {
	t := fasttemplate.New(template, "{{", "}}")
	return t.ExecuteString(param)
}
