package support

import "math"

// Page 是分页信息
type Page struct {
	// 页码
	Page int `json:"page"`
	// 每页条数
	Size int `json:"size"`
	// 总数
	Count int64 `json:"count"`
}

// PageCount 返回页数
func (pg Page) PageCount() int {
	f := math.Ceil(float64(pg.Count) / float64(pg.Size))
	return int(f)
}

type BaseQueryPageCmd struct {
	Page int `query:"page"`
	Size int `query:"size"`
}

func (cmd BaseQueryPageCmd) Offset() int {
	return (cmd.Page - 1) * cmd.Size
}
