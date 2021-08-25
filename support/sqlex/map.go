package sqlex

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

// Map 是map
type Map map[string]interface{}

// Value 实现 driver.Valuer
func (m Map) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan 实现 driver.Scaner
func (m *Map) Scan(value interface{}) error {
	var v []byte
	switch t := value.(type) {
	case []uint8:
		v = []byte(t)
		break
	case string:
		v = []byte(t)
		break
	default:
		return errors.New("unknown type in Map.Scan")
	}

	if v == nil || len(v) == 0 {
		*m = Map{}
		return nil
	}

	return json.Unmarshal(v, m)
}

func (m *Map) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "" || s == "{}" {
		*m = map[string]interface{}{}
	} else {
		if strings.HasPrefix(s, "\"") {
			// 从binlog里解析时，params是"{}"，需要先json.Unmarshal一次
			var s2 string
			if err := json.Unmarshal(b, &s2); err != nil {
				return err
			}
			var mp map[string]interface{}
			if err := json.Unmarshal([]byte(s2), &mp); err != nil {
				return err
			}
			*m = mp
		} else {
			var mp map[string]interface{}
			if err := json.Unmarshal(b, &mp); err != nil {
				return err
			}
			*m = mp
		}
	}
	return nil
}

// func (m Map) MarshalJSON() ([]byte, error) {
// 	s := strings.Join(il, ",")
// 	return []byte(s), nil
// }
