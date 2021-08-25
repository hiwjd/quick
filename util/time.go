package util

import (
	"time"
)

// DiffDays 计算t1,t2之间的天差
func DiffDays(t1 time.Time, t2 time.Time) int {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())
	if t1.After(t2) {
		t1, t2 = t2, t1
	}
	df := t2.Sub(t1).Hours() / 24
	return int(df)
}

// BeginningOfMonth 获取t月第一天
func BeginningOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 获取t月最后一天
func EndOfMonth(t time.Time) time.Time {
	beginning := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	day := beginning.AddDate(0, 1, -1).Day()
	return time.Date(t.Year(), t.Month(), day, 23, 59, 59, 0, t.Location())
}

// BeginningOfDay 获取t日起始
func BeginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取t日结束
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}
