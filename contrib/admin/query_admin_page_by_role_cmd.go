package admin

import "gorm.io/gorm"

// QueryAdminPageByRoleCmd 是根据角色查询管理员的命令
type QueryAdminPageByRoleCmd struct {
	Page       int    `query:"page"`
	Size       int    `query:"size"`
	RoleIDList []uint `query:"roleIDList"`
	Account    string `query:"account"`
	Name       string `query:"name"`
	Mobile     string `query:"mobile"`
	Active     *bool  `query:"active"`
}

func (cmd QueryAdminPageByRoleCmd) offset() int {
	return (cmd.Page - 1) * cmd.Size
}

func (cmd QueryAdminPageByRoleCmd) applyCondition(db *gorm.DB) *gorm.DB {
	if cmd.RoleIDList != nil {
		db = db.Where("admin_role.role_id in (?)", cmd.RoleIDList)
	}
	if cmd.Account != "" {
		db = db.Where("admin.account = ?", cmd.Account)
	}
	if cmd.Name != "" {
		db = db.Where("admin.name = ?", cmd.Name)
	}
	if cmd.Mobile != "" {
		db = db.Where("admin.mobile = ?", cmd.Mobile)
	}
	if cmd.Active != nil {
		db = db.Where("admin.active = ?", cmd.Active)
	}
	return db
}
