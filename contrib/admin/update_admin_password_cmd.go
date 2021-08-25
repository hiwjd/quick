package admin

// UpdateAdminPasswordCmd 是修改管理员密码的命令
type UpdateAdminPasswordCmd struct {
	ID       uint   `json:"id" validate:"nonzero"`
	Origin   string `json:"origin" vaidate:"nonzero"`
	Password string `json:"password" validate:"nonzero"`
}
