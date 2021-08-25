package sqlex

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

// StringList 是string数组
type StringList []string

// Value 实现 driver.Valuer
func (il StringList) Value() (driver.Value, error) {
	return strings.Join(il, ","), nil
}

// Scan 实现 driver.Scaner
func (il *StringList) Scan(value interface{}) error {
	var v string
	switch t := value.(type) {
	case []uint8:
		v = string(t)
		break
	case string:
		v = t
		break
	default:
		return errors.New("unknown type in StringList.Scan")
	}
	v = strings.Trim(v, " ")
	// v := value.(string)
	if v == "" {
		*il = StringList{}
	} else {
		*il = StringList(strings.Split(v, ","))
	}
	return nil
}

func (il *StringList) UnmarshalJSON(b []byte) (err error) {
	isJSONFmt := bytes.HasPrefix(b, []byte("["))
	if isJSONFmt {
		var arr []string
		if err = json.Unmarshal(b, &arr); err != nil {
			return
		}
		*il = arr
	} else {
		s := string(b)
		*il = StringList(strings.Split(s, ","))
	}
	return nil
}

// func (il StringList) MarshalJSON() ([]byte, error) {
// 	s := strings.Join(il, ",")
// 	return []byte(s), nil
// }
