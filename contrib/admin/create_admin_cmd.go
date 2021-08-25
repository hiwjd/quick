package admin

import "github.com/rs/xid"

// CreateAdminCmd 是添加管理员的命令
type CreateAdminCmd struct {
	Account    string `json:"account" validate:"nonzero"`
	Password   string `json:"password" validate:"nonzero"`
	Name       string `json:"name" validate:"nonzero"`
	Mobile     string `json:"mobile" validate:"nonzero"`
	Active     bool   `json:"active"`
	RoleIDList []uint `json:"roleIdList"`
}

func (cmd *CreateAdminCmd) toModel() (*Admin, []*AdminRole, error) {
	arList := make([]*AdminRole, len(cmd.RoleIDList))
	for i, roleID := range cmd.RoleIDList {
		arList[i] = &AdminRole{
			RoleID: roleID,
		}
	}

	model := &Admin{
		Account: cmd.Account,
		Name:    cmd.Name,
		Mobile:  cmd.Mobile,
		Active:  cmd.Active,
		Code:    xid.New().String(),
	}
	model.SetPassword(cmd.Password)

	return model, arList, nil
}
