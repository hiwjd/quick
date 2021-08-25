package admin

import "gorm.io/gorm"

// QueryAdminPageCmd 是查询管理员分页列表的命令
type QueryAdminPageCmd struct {
	Page    int    `query:"page"`
	Size    int    `query:"size"`
	Account string `query:"account"`
	Name    string `query:"name"`
	Mobile  string `query:"mobile"`
	Active  *bool  `query:"active"`
}

func (cmd QueryAdminPageCmd) offset() int {
	return (cmd.Page - 1) * cmd.Size
}

func (cmd QueryAdminPageCmd) applyCondition(db *gorm.DB) *gorm.DB {
	if cmd.Account != "" {
		db = db.Where("account = ?", cmd.Account)
	}
	if cmd.Name != "" {
		db = db.Where("name = ?", cmd.Name)
	}
	if cmd.Mobile != "" {
		db = db.Where("mobile = ?", cmd.Mobile)
	}
	if cmd.Active != nil {
		db = db.Where("active = ?", cmd.Active)
	}
	return db
}
