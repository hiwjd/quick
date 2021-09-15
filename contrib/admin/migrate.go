package admin

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		Admin{},
		AdminRole{},
		Role{},
		RoleMenu{},
		Menu{},
		API{},
		WxAdmin{},
	)
}
