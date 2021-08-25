package quick

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type forTest struct {
	ID       uint   `validate:"nonzero" title:"编号"`
	Name     string `validate:"nonzero,max=5" title:"姓名"`
	Password string `validate:"nonzero,min=6,max=8" title:"密码"`
}

func TestExtractTitleMap(t *testing.T) {
	v := forTest{}
	err := Check(v)
	assert.NotNil(t, err)

	errArr, ok := err.(ErrorArray)
	assert.True(t, ok)
	assert.NotNil(t, errArr)
	assert.Equal(t, 3, len(errArr))
	assert.Equal(t, ErrorData{Field: "ID", Title: "编号", Errors: []string{"不能为空"}}, errArr[0])
	assert.Equal(t, ErrorData{Field: "Name", Title: "姓名", Errors: []string{"不能为空"}}, errArr[1])
	assert.Equal(t, ErrorData{Field: "Password", Title: "密码", Errors: []string{"不能为空", "太小了"}}, errArr[2])
	assert.Equal(t, "编号: 不能为空, 姓名: 不能为空, 密码: 不能为空, 太小了", errArr.Error())
}
