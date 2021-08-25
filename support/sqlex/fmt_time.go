package sqlex

import (
	"database/sql/driver"
	"errors"
	"time"
)

// FMTime 主要是定制了json.MarshalJSON接口
type FMTime time.Time

// Value 实现 driver.Valuer
func (ft FMTime) Value() (driver.Value, error) {
	return time.Time(ft), nil
}

// Scan 实现 driver.Scaner
func (ft *FMTime) Scan(value interface{}) error {
	switch t := value.(type) {
	case time.Time:
		*ft = FMTime(t)
		break
	default:
		return errors.New("unknown type in FMTime.Scan")
	}
	return nil
}

// MarshalJSON 实现json.MarshalJSON
func (ft FMTime) MarshalJSON() ([]byte, error) {
	t := time.Time(ft)
	if t.IsZero() {
		return []byte(`""`), nil
	}
	s := t.Format(`"2006-01-02 15:04:05"`)
	return []byte(s), nil
}

func (ft FMTime) Format(fmt string) string {
	return time.Time(ft).Format(fmt)
}

func (ft FMTime) IsZero() bool {
	return time.Time(ft).IsZero()
}
