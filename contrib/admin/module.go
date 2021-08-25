package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hiwjd/quick"
	"github.com/hiwjd/quick/support/session"
	"github.com/hiwjd/quick/util"
	"github.com/labstack/echo/v4"
)

const AdminSessionID = "AdminSessionID"

// Session 是登录信息
type Session struct {
	ID       uint // 管理员ID
	Name     string
	Mobile   string
	Remember bool
}

func AdminModule(ac quick.AppContext) {
	adminService := ac.Take("admin").(Service)
	adminSessionStorage := ac.Take("adminSessionStorage").(session.Storage)
	ct := &ctrl{
		adminService:        adminService,
		adminSessionStorage: adminSessionStorage,
	}
	ac.POST("/pub/admin/login", ct.adminLogin)                     // 后台 - 登录
	ac.POST("/ana/admin/logout", ct.adminLogout)                   // 后台 - 登出
	ac.POST("/ana/admin/update-my-pass", ct.adminUpdateMyPassword) // 后台 - 修改自己的密码
	ac.GET("/ana/admin/menu", ct.queryAdminMenu)                   // 后台 - 当前登录管理员的菜单
	ac.GET("/ana/admin/query-admin-page", func(c echo.Context) error {
		xx := c.QueryParam("xx")
		ac.Publish("admin-query", xx)
		return ct.queryAdminPage(c)
	}) // 后台 - 管理员分页列表
	ac.GET("/ana/admin/get-by-id", ct.getAdminByID)                   // 后台 - 根据ID查询管理员
	ac.GET("/ana/admin/get-by-account", ct.getAdminByAccount)         // 后台 - 根据帐号查询管理员
	ac.POST("/ana/admin/create", ct.createAdmin)                      // 后台 - 创建管理员
	ac.POST("/ana/admin/update", ct.updateAdmin)                      // 后台 - 更新管理员
	ac.POST("/ana/admin/update-password", ct.updateAdminPassword)     // 后台 - 更新管理员密码
	ac.GET("/ana/admin/query-role-list", ct.queryRoleList)            // 后台 - 角色列表
	ac.GET("/ana/admin/query-admin-role-list", ct.queryAdminRoleList) // 后台 - 管理员的角色列表
	// ac.GET("/ana/admin/get-qiniu-upload-token", ct.getQiniuUploadToken) // 后台 - 获取上传七牛云的token
	// ac.POST("/ana/admin/upload", ct.upload)                             // 后台 - 上传图片
	// ac.Static("/pub/fread", conf.UploadDir)
	ac.Subscribe("admin-query", func(s string) {
		fmt.Printf(">>> admin-query-s1: %s\n", s)
	})
	ac.Subscribe("admin-query", func(s string) {
		time.Sleep(time.Second * 2)
		fmt.Printf(">>> admin-query-s2: %s\n", s)
		if n, er := strconv.Atoi(s); er != nil {
			ac.Logf("er: %s\n", er.Error())
		} else {
			n2 := 1 / n
			ac.Logf("n2: %d\n", n2)
		}
	})
}

// AdminLoginReq 是管理员登录请求
type AdminLoginReq struct {
	Account  string `json:"account" validate:"nonzero"`
	Password string `json:"password" validate:"nonzero"`
	Remember bool   `json:"remember"`
}

type ctrl struct {
	adminService        Service
	adminSessionStorage session.Storage
}

func (ct *ctrl) adminLogin(c echo.Context) (err error) {
	ctx := c.Request().Context()
	req := new(AdminLoginReq)
	if err = c.Bind(req); err != nil {
		return
	}
	if err = c.Validate(req); err != nil {
		return
	}

	var admin *Admin
	if admin, err = ct.adminService.GetAdminByAccount(ctx, req.Account); err != nil {
		err = echo.NewHTTPError(http.StatusBadRequest, "帐号或密码错误").SetInternal(err)
		return
	}

	if !admin.CheckPassword(req.Password) {
		err = echo.NewHTTPError(http.StatusBadRequest, "帐号或密码错误")
		return
	}

	ttl := 30 * time.Minute
	if req.Remember {
		ttl = 0
	}

	session := &Session{
		ID:       admin.ID,
		Name:     admin.Name,
		Mobile:   admin.Mobile,
		Remember: req.Remember,
	}

	token, err := ct.adminSessionStorage.Set(session, ttl)
	if err != nil {
		err = echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
		return
	}

	return c.JSON(http.StatusOK, util.Map{"token": token, "data": nil})
}

func (ct *ctrl) queryAdminMenu(c echo.Context) (err error) {
	ctx := c.Request().Context()
	session := c.Get(AdminSessionID).(Session)

	var menuList []*Menu
	if menuList, err = ct.adminService.QueryMenuListByAdminID(ctx, session.ID); err != nil {
		return
	}

	ml := MenuList(menuList)

	return c.JSON(http.StatusOK, util.Map{"menu": ml.AsNode()})
}

func (ct *ctrl) adminLogout(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (ct *ctrl) queryAdminPage(c echo.Context) (err error) {
	ctx := c.Request().Context()
	cmd := &QueryAdminPageCmd{
		Page: 1,
		Size: 20,
	}
	if err = c.Bind(cmd); err != nil {
		return
	}

	rows, pg, err := ct.adminService.QueryAdminPage(ctx, cmd)
	if err != nil {
		return
	}

	return c.JSON(http.StatusOK, util.Map{"data": rows, "pg": pg})
}

func (ct *ctrl) getAdminByID(c echo.Context) (err error) {
	ctx := c.Request().Context()
	s := c.QueryParam("id")
	id, err := strconv.ParseUint(s, 10, 32)
	admin, err := ct.adminService.GetAdminByID(ctx, uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, nil)
	}

	return c.JSON(http.StatusOK, admin)
}

func (ct *ctrl) getAdminByAccount(c echo.Context) (err error) {
	ctx := c.Request().Context()
	account := c.QueryParam("account")
	admin, err := ct.adminService.GetAdminByAccount(ctx, account)
	if err != nil {
		return c.JSON(http.StatusNotFound, util.Map{"message": "maybe it`s ok?"})
	}

	return c.JSON(http.StatusOK, admin)
}

func (ct *ctrl) createAdmin(c echo.Context) (err error) {
	ctx := c.Request().Context()
	cmd := new(CreateAdminCmd)
	if err = c.Bind(cmd); err != nil {
		return
	}
	if err = c.Validate(cmd); err != nil {
		return
	}

	var admin *Admin
	if admin, err = ct.adminService.CreateAdmin(ctx, cmd); err != nil {
		return
	}

	return c.JSON(http.StatusOK, admin)
}

func (ct *ctrl) updateAdmin(c echo.Context) (err error) {
	ctx := c.Request().Context()
	cmd := new(UpdateAdminCmd)
	if err = c.Bind(cmd); err != nil {
		return
	}

	if err = ct.adminService.UpdateAdmin(ctx, cmd); err != nil {
		return
	}

	return c.JSON(http.StatusOK, util.Map{"message": "更新成功"})
}

func (ct *ctrl) queryRoleList(c echo.Context) (err error) {
	ctx := c.Request().Context()
	cmd := new(QueryRoleListCmd)
	if err = c.Bind(cmd); err != nil {
		return
	}

	var roleList []*Role
	if roleList, err = ct.adminService.QueryRoleList(ctx, cmd); err != nil {
		return
	}

	return c.JSON(http.StatusOK, roleList)
}

func (ct *ctrl) queryAdminRoleList(c echo.Context) (err error) {
	ctx := c.Request().Context()
	s := c.QueryParam("id")
	id, err := strconv.ParseUint(s, 10, 32)

	var roleIDList []uint
	if roleIDList, err = ct.adminService.QueryRoleIDListByAdminID(ctx, uint(id)); err != nil {
		return
	}

	return c.JSON(http.StatusOK, roleIDList)
}

func (ct *ctrl) updateAdminPassword(c echo.Context) (err error) {
	ctx := c.Request().Context()
	cmd := new(UpdateAdminPasswordCmd)
	if err = c.Bind(cmd); err != nil {
		return
	}
	if err = c.Validate(cmd); err != nil {
		return
	}

	if err = ct.adminService.UpdateAdminPassword(ctx, cmd); err != nil {
		return
	}

	return c.JSON(http.StatusOK, util.Map{"message": "修改成功"})
}

func (ct *ctrl) getAdminByCode(c echo.Context) (err error) {
	ctx := c.Request().Context()
	code := c.QueryParam("code")
	admin, err := ct.adminService.GetAdminByCode(ctx, code)
	if err != nil {
		return
	}

	admin.Desensitive()

	return c.JSON(http.StatusOK, admin)
}

func (ct *ctrl) adminUpdateMyPassword(c echo.Context) (err error) {
	ctx := c.Request().Context()
	session := c.Get(AdminSessionID).(Session)

	cmd := new(UpdateAdminPasswordCmd)
	if err = c.Bind(cmd); err != nil {
		return
	}
	cmd.ID = session.ID

	if err = ct.adminService.UpdateAdminPassword(ctx, cmd); err != nil {
		return
	}

	return c.JSON(http.StatusOK, util.Map{"message": "修改成功"})
}
