package admin

import (
	"context"
	"net/http"

	"github.com/hiwjd/quick"
	"github.com/hiwjd/quick/support"
	"gorm.io/gorm"
)

// Service 是管理员服务
type Service interface {
	QueryAdminPage(context.Context, *QueryAdminPageCmd) ([]*Admin, *support.Page, error)                                 // 查询管理员分页列表
	GetAdminByID(context.Context, uint) (*Admin, error)                                                                  // 根据ID查找管理员
	GetAdminByAccount(context.Context, string) (*Admin, error)                                                           // 根据帐号查找管理员
	GetAdminByWxopenid(context.Context, string) (*Admin, error)                                                          // 根据微信openid查找管理员
	CreateAdmin(context.Context, *CreateAdminCmd) (*Admin, error)                                                        // 创建管理员
	UpdateAdmin(context.Context, *UpdateAdminCmd) error                                                                  // 修改管理员
	UpdateAdminPassword(context.Context, *UpdateAdminPasswordCmd) error                                                  // 修改管理员密码
	QueryRoleListByAdminID(context.Context, uint) (Roles, error)                                                         // 查询管理员的角色列表
	QueryRoleIDListByAdminID(context.Context, uint) ([]uint, error)                                                      // 查询管理员的角色ID列表
	QueryMenuListByAdminID(context.Context, uint) ([]*Menu, error)                                                       // 查询管理员拥有的菜单列表
	QueryRoleList(context.Context, *QueryRoleListCmd) ([]*Role, error)                                                   // 查询所有角色列表
	GetAdminByCode(ctx context.Context, code string) (data *Admin, err error)                                            // 根据code查找管理员
	BindMiniapp(ctx context.Context, code, openid, password string) (err error)                                          // 管理员绑定小程序
	QueryAdminPageByRole(ctx context.Context, cmd *QueryAdminPageByRoleCmd) (data []*Admin, pg *support.Page, err error) // 根据角色查询管理员
	CanAccessAPI(ctx context.Context, adminID uint, method, path string) bool                                            // 返回能否访问某个接口
	GetPermData(ctx context.Context, adminID uint) []DataPerm                                                            // 返回管理员的数据权限
}

// NewService 构造管理员服务
func NewService(db *gorm.DB) Service {
	return &service{
		db: db,
	}
}

type service struct {
	db *gorm.DB
}

func (s *service) QueryAdminPage(ctx context.Context, cmd *QueryAdminPageCmd) (data []*Admin, pg *support.Page, err error) {
	db := s.db.WithContext(ctx)

	pg = &support.Page{}
	pg.Page = cmd.Page
	pg.Size = cmd.Size
	db = cmd.applyCondition(db.Model(&Admin{}))
	err = db.Count(&pg.Count).Limit(cmd.Size).Offset(cmd.offset()).Find(&data).Error

	return
}

func (s *service) GetAdminByAccount(ctx context.Context, account string) (*Admin, error) {
	var admin Admin
	err := s.db.Where("account = ?", account).First(&admin).Error
	return &admin, err
}

func (s *service) GetAdminByID(ctx context.Context, id uint) (admin *Admin, err error) {
	admin = new(Admin)
	err = s.db.Where("id=?", id).First(admin).Error
	return
}

func (s *service) GetAdminByWxopenid(ctx context.Context, openid string) (admin *Admin, err error) {
	wxadmin := new(WxAdmin)
	if err = s.db.Where("openid = ?", openid).Take(&wxadmin).Error; err != nil {
		return
	}

	return s.GetAdminByID(ctx, wxadmin.AdminID)
}

func (s *service) CreateAdmin(ctx context.Context, cmd *CreateAdminCmd) (admin *Admin, err error) {
	var arList []*AdminRole
	if admin, arList, err = cmd.toModel(); err != nil {
		return
	}

	err = s.db.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(admin).Error; err != nil {
			return
		}

		for _, ar := range arList {
			ar.AdminID = admin.ID
			if err = tx.Create(ar).Error; err != nil {
				return
			}
		}

		return
	})

	return
}

func (s *service) UpdateAdmin(ctx context.Context, cmd *UpdateAdminCmd) (err error) {
	var admin *Admin
	var arList []*AdminRole
	if admin, arList, err = cmd.toModel(); err != nil {
		return
	}

	err = s.db.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Model(admin).Updates(admin).Error; err != nil {
			return
		}

		tx.Where("admin_id = ?", admin.ID).Delete(AdminRole{})
		for _, ar := range arList {
			ar.AdminID = admin.ID
			if err = tx.Create(ar).Error; err != nil {
				return
			}
		}

		return
	})

	return
}

func (s *service) QueryRoleListByAdminID(ctx context.Context, adminID uint) (data Roles, err error) {
	var arList []*AdminRole
	if err = s.db.Where("admin_id = ?", adminID).Find(&arList).Error; err != nil {
		return
	}

	roleIDList := make([]uint, len(arList))
	for i, ar := range arList {
		roleIDList[i] = ar.RoleID
	}

	err = s.db.Where("id in (?)", roleIDList).Find(&data).Error
	return
}

func (s *service) QueryRoleIDListByAdminID(ctx context.Context, adminID uint) (roleIDList []uint, err error) {
	var arList []*AdminRole
	if err = s.db.Where("admin_id = ?", adminID).Find(&arList).Error; err != nil {
		return
	}

	roleIDList = make([]uint, len(arList))
	for i, ar := range arList {
		roleIDList[i] = ar.RoleID
	}

	return
}

func (s *service) QueryMenuListByAdminID(ctx context.Context, adminID uint) (menuList []*Menu, err error) {
	var roleIDList []uint
	if roleIDList, err = s.QueryRoleIDListByAdminID(ctx, adminID); err != nil {
		return
	}

	var rmList []*RoleMenu
	if err = s.db.Where("role_id in (?)", roleIDList).Find(&rmList).Error; err != nil {
		return
	}

	var menuIDList []uint
	for _, rm := range rmList {
		menuIDList = append(menuIDList, rm.MenuID)
	}

	err = s.db.Where("id in (?)", menuIDList).Find(&menuList).Error
	return
}

func (s *service) QueryRoleList(ctx context.Context, cmd *QueryRoleListCmd) (roleList []*Role, err error) {
	db := cmd.applyCondition(s.db.Model(&Role{}))

	err = db.Find(&roleList).Error
	return
}

func (s *service) UpdateAdminPassword(ctx context.Context, cmd *UpdateAdminPasswordCmd) (err error) {
	if err = quick.Check(cmd); err != nil {
		return
	}

	adm, err := s.GetAdminByID(ctx, cmd.ID)
	if err != nil {
		return
	}

	if !adm.CheckPassword(cmd.Origin) {
		err = support.NewFineErr(http.StatusBadRequest, "原密码错误")
		return
	}

	admin := Admin{
		ID: cmd.ID,
	}
	admin.SetPassword(cmd.Password)

	err = s.db.Model(admin).Updates(admin).Error
	return
}

func (s *service) GetAdminByCode(ctx context.Context, code string) (data *Admin, err error) {
	data = new(Admin)
	err = s.db.Where("code = ?", code).Take(data).Error
	return
}

func (s *service) BindMiniapp(ctx context.Context, code, openid, password string) (err error) {
	admin, err := s.GetAdminByCode(ctx, code)
	if err != nil {
		return
	}

	if !admin.CheckPassword(password) {
		err = support.NewFineErr(http.StatusBadRequest, "密码错误")
		return
	}

	var wa WxAdmin
	err = s.db.Where("openid = ?", openid).Take(&wa).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}

	if err == gorm.ErrRecordNotFound {
		wa = WxAdmin{
			Openid:  openid,
			AdminID: admin.ID,
		}
		err = s.db.Create(&wa).Error
	} else {
		// 更新
		wa.AdminID = admin.ID
		err = s.db.Save(&wa).Error
	}
	return
}

func (s *service) QueryAdminPageByRole(ctx context.Context, cmd *QueryAdminPageByRoleCmd) (data []*Admin, pg *support.Page, err error) {
	pg = &support.Page{}
	pg.Page = cmd.Page
	pg.Size = cmd.Size
	db := s.db.Table("admin").Select("admin.*").Joins("left join admin_role on admin.id = admin_role.admin_id")
	db = cmd.applyCondition(db)
	err = db.Count(&pg.Count).Limit(cmd.Size).Offset(cmd.offset()).Find(&data).Error
	return
}

func (s *service) queryAPIsByAdminID(ctx context.Context, adminID uint) map[string]bool {
	m := make(map[string]bool, 0)
	var roleIDs []uint
	if err := s.db.Model(AdminRole{}).Where("admin_id = ?", adminID).Pluck("role_id", &roleIDs).Error; err != nil || len(roleIDs) < 1 {
		return m
	}

	var menuIDs []uint
	if err := s.db.Model(RoleMenu{}).Where("role_id IN (?)", roleIDs).Pluck("menu_id", &menuIDs).Error; err != nil || len(menuIDs) < 1 {
		return m
	}

	// 9999 包含公有的接口
	menuIDs = append(menuIDs, 9999)

	var menus []Menu
	if err := s.db.Model(Menu{}).Where("id IN (?)", menuIDs).Find(&menus).Error; err != nil {
		return m
	}

	for _, menu := range menus {
		for _, api := range menu.APIList {
			m[api] = true
		}
	}

	return m
}

func (s *service) CanAccessAPI(ctx context.Context, adminID uint, method, path string) bool {
	m := s.queryAPIsByAdminID(ctx, adminID)
	_, ok := m[method+path]
	return ok
}

func (s *service) GetPermData(ctx context.Context, adminID uint) []DataPerm {
	var ars []AdminRole
	if err := s.db.Where("admin_id = ?", adminID).Find(&ars).Error; err != nil {
		return []DataPerm{}
	}

	roleIDs := make([]uint, len(ars))
	for i, ar := range ars {
		roleIDs[i] = ar.RoleID
	}

	if len(roleIDs) < 1 {
		return []DataPerm{}
	}

	var roles []Role
	if err := s.db.Where("id IN (?)", roleIDs).Find(&roles).Error; err != nil {
		return []DataPerm{}
	}

	var dps []DataPerm
	for _, role := range roles {
		for _, dp := range role.DataPerms {
			dps = append(dps, dp)
		}
	}

	return dps
}
