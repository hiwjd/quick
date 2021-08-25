package admin

import "gorm.io/gorm"

// QueryRoleListCmd 查询角色列表的命令
type QueryRoleListCmd struct {
	Group string `query:"group"`
	Key   string `query:"key"`
}

func (cmd QueryRoleListCmd) applyCondition(db *gorm.DB) *gorm.DB {
	if cmd.Group != "" {
		db = db.Where("`group` = ?", cmd.Group)
	}
	if cmd.Key != "" {
		db = db.Where("`key` = ?", cmd.Key)
	}
	return db
}
