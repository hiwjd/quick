package admin

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type DataPerm struct {
	Domain string                     `json:"domain"`
	Perm   map[string]json.RawMessage `json:"perm"`
}
type DataPerms []DataPerm

// Value 实现driver.Valuer
func (r DataPerms) Value() (driver.Value, error) {
	bs, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return string(bs), nil
}

// Scan 实现 driver.Scaner
func (r *DataPerms) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), r)
	case []uint8:
		return json.Unmarshal(v, r)
	default:
		return fmt.Errorf("unknown type in DataPerms.Scan")
	}
}
