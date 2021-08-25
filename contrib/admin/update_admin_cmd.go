package admin

// UpdateAdminCmd 是修改管理员的命令
type UpdateAdminCmd struct {
	ID         uint   `json:"id" validate:"nonzero"`
	Account    string `json:"account" validate:"nonzero"`
	Name       string `json:"name" validate:"nonzero"`
	Mobile     string `json:"mobile" validate:"nonzero"`
	Active     bool   `json:"active"`
	RoleIDList []uint `json:"roleIdList"`
}

func (cmd *UpdateAdminCmd) toModel() (*Admin, []*AdminRole, error) {
	arList := make([]*AdminRole, len(cmd.RoleIDList))
	for i, roleID := range cmd.RoleIDList {
		arList[i] = &AdminRole{
			RoleID: roleID,
		}
	}

	model := &Admin{
		ID:      cmd.ID,
		Account: cmd.Account,
		Name:    cmd.Name,
		Mobile:  cmd.Mobile,
		Active:  cmd.Active,
	}

	return model, arList, nil
}
