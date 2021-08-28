# admin

> 管理员模块，包括账号、角色、权限的管理接口；账号登录、修改密码的接口以及接口访问权限检查

## 依赖

- [adminSessionStorage](github.com/hiwjd/quick/blob/main/support/session/storage.go)

## Provider

- [adminService](github.com/hiwjd/quick/blob/main/contrib/admin/service.go)

## HTTP接口

- POST `/pub/admin/login` 账号登录
- POST `/ana/admin/logout` 账号登出
- POST `/ana/admin/update-my-pass` 修改当前会话账号的密码
- GET `/ana/admin/menu` 查询当前会话账号的菜单
- GET `/ana/admin/query-admin-page` 查询账号列表
- GET `/ana/admin/get-by-id"` 根据ID获取账号信息
- GET `/ana/admin/get-by-account` 根据账号获取账号信息
- POST `/ana/admin/create` 添加账号
- POST `/ana/admin/update` 修改账号
- POST `/ana/admin/update-password` 修改指定账号的密码
- GET `/ana/admin/query-role-list` 查询角色列表
- GET `/ana/admin/query-admin-role-list` 查询账号的角色
