package admin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAdmin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	assert.Nil(t, err)

	db.AutoMigrate(Admin{}, Role{}, AdminRole{}, Menu{}, API{})

	service := NewService(db)
	ctx := context.Background()

	cmd := &CreateAdminCmd{
		Account:    "admin",
		Password:   "123123",
		Name:       "管理员",
		Active:     true,
		RoleIDList: []uint{1, 2},
	}
	admin, err := service.CreateAdmin(ctx, cmd)
	assert.Nil(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, cmd.Account, admin.Account)
	assert.True(t, admin.CheckPassword(cmd.Password))
	assert.Equal(t, cmd.Name, admin.Name)
	assert.Equal(t, cmd.Active, admin.Active)

	roleIDList, err := service.QueryRoleIDListByAdminID(ctx, admin.ID)
	assert.Nil(t, err)
	assert.Equal(t, cmd.RoleIDList, roleIDList)
}
