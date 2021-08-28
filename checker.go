package quick

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/validator.v2"
)

var (
	// ErrUnsupported 表示Checker遇到了不支持的情况
	ErrUnsupported = errors.New("Checker not support this type")
)

// Checker 校验器
type Checker interface {
	// Check 校验v，并返回结构化的错误信息ErrorArray或者其他error
	Check(v interface{}) error
}

var defaultsErrorMsgMap = map[string]string{
	"zeroValue":    "不能为空",
	"min":          "太小了",
	"max":          "太大了",
	"len":          "长度不正确",
	"regexp":       "不符合规则",
	"unsupport":    "类型错误",
	"badParameter": "类型错误",
}

// dftChecker 是默认的Checker
var dftChecker = NewChecker("validate", "title", nil)

// Check 调用默认实现的Check
func Check(v interface{}) error {
	return dftChecker.Check(v)
}

// New 返回Checker实例
func NewChecker(ruleTagName, titleTagName string, errorMsgMap map[string]string) Checker {
	emm := defaultsErrorMsgMap
	if errorMsgMap != nil {
		for k, v := range errorMsgMap {
			emm[k] = v
		}
	}

	validator := validator.NewValidator()
	validator.SetTag(ruleTagName)

	return &checker{
		ruleTagName:  ruleTagName,
		titleTagName: titleTagName,
		errorMsgMap:  emm,
		validator:    validator,
	}
}

// ErrorData 是某个字段的校验错误
type ErrorData struct {
	Field  string   `json:"field"`  // 字段
	Title  string   `json:"title"`  // 字段标题
	Errors []string `json:"errors"` // 错误
}

// ErrorArray 是字段的错误信息
type ErrorArray []ErrorData

func (eArr ErrorArray) Error() string {
	var b bytes.Buffer

	for _, ed := range eArr {
		if len(ed.Errors) > 0 {
			b.WriteString(fmt.Sprintf("%s: ", ed.Title))
			for _, err := range ed.Errors {
				b.WriteString(fmt.Sprintf("%s, ", err))
			}
		}
	}

	return strings.TrimSuffix(b.String(), ", ")
}

type checker struct {
	ruleTagName  string
	titleTagName string
	errorMsgMap  map[string]string
	validator    *validator.Validator
}

// Check 检查struct并返回友好的提示
func (c *checker) Check(v interface{}) (err error) {
	var titleMap map[string]string
	if titleMap, err = c.extractStructFields(v); err != nil {
		return err
	}

	getTitle := func(field string) string {
		if title, ok := titleMap[field]; ok && title != "" {
			return title
		}
		return field
	}

	err = c.validator.Validate(v)
	if err == nil {
		return nil
	}

	var errArr []ErrorData
	if errMap, ok := err.(validator.ErrorMap); ok {
		for k, eArr := range errMap {
			ed := ErrorData{
				Field:  k,
				Title:  getTitle(k),
				Errors: c.fmtErrArr(eArr),
			}
			errArr = append(errArr, ed)
		}
	}
	return ErrorArray(errArr)
}

func (c checker) fmtErrArr(errArr validator.ErrorArray) []string {
	ss := []string{}
	for _, er := range errArr {
		switch er {
		case validator.ErrZeroValue:
			ss = append(ss, c.errorMsgMap["zeroValue"])
			break
		case validator.ErrMin:
			ss = append(ss, c.errorMsgMap["min"])
			break
		case validator.ErrMax:
			ss = append(ss, c.errorMsgMap["max"])
			break
		case validator.ErrLen:
			ss = append(ss, c.errorMsgMap["len"])
			break
		case validator.ErrRegexp:
			ss = append(ss, c.errorMsgMap["regexp"])
			break
		case validator.ErrUnsupported:
			ss = append(ss, c.errorMsgMap["unsupport"])
			break
		case validator.ErrBadParameter:
			ss = append(ss, c.errorMsgMap["badParameter"])
			break
		}
	}
	return ss
}

// extractStructFields 从struct中提取标题，组成以字段名为键标题为值的map返回
func (c checker) extractStructFields(v interface{}) (map[string]string, error) {
	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	if kind != reflect.Struct && kind != reflect.Ptr {
		return nil, ErrUnsupported
	}

	if kind == reflect.Ptr {
		if rv.Elem().Kind() != reflect.Struct {
			return nil, ErrUnsupported
		}
		rv = rv.Elem()
	}

	st := rv.Type()
	nfields := st.NumField()

	m := make(map[string]string, nfields)
	for i := 0; i < nfields; i++ {
		sf := st.Field(i)
		title := sf.Tag.Get(c.titleTagName)
		m[sf.Name] = title
	}

	return m, nil
}
