package sqlex

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

// UintList 是uint数组
type UintList []uint

// Value 实现 driver.Valuer
func (il UintList) Value() (driver.Value, error) {
	return json.Marshal(il)
}

// Scan 实现 driver.Scaner
func (il *UintList) Scan(value interface{}) error {
	var v []byte
	switch t := value.(type) {
	case []uint8:
		v = []byte(t)
		break
	case string:
		v = []byte(t)
		break
	default:
		return errors.New("unknown type in UintList.Scan")
	}

	if v == nil || len(v) == 0 {
		*il = UintList{}
		return nil
	}

	return json.Unmarshal(v, il)
}

func (il *UintList) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	if s == "" || s == "{}" {
		*il = []uint{}
	} else {
		if strings.HasPrefix(s, "\"") {
			// 从binlog里解析时，params是"{}"，需要先json.Unmarshal一次
			var s2 string
			if err := json.Unmarshal(b, &s2); err != nil {
				return err
			}
			var mp []uint
			if err := json.Unmarshal([]byte(s2), &mp); err != nil {
				return err
			}
			*il = mp
		} else {
			var mp []uint
			if err := json.Unmarshal(b, &mp); err != nil {
				return err
			}
			*il = mp
		}
	}
	return nil
}
