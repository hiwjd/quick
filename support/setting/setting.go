package setting

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Setting interface {
	Get(key string) (string, error)
	GetObj(key string, v interface{}) error
	Set(key, value string) error
	SetObj(key string, v interface{}) error
}

func New(db *gorm.DB) Setting {
	return &dbSetting{
		db: db,
	}
}

type dbSetting struct {
	db *gorm.DB
}

func (ds *dbSetting) Get(id string) (string, error) {
	var m SettingModel
	if err := ds.db.Where("id = ?", id).Take(&m).Error; err != nil {
		return "", err
	}

	return m.Value, nil
}

func (ds *dbSetting) GetObj(id string, v interface{}) error {
	var m SettingModel
	if err := ds.db.Where("id = ?", id).Take(&m).Error; err != nil {
		return err
	}

	return json.Unmarshal([]byte(m.Value), v)
}

func (ds *dbSetting) Set(id string, v string) error {
	return ds.db.Model(SettingModel{}).Where("id = ?", id).Update("value", v).Error
}

func (ds *dbSetting) SetObj(id string, v interface{}) (err error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return
	}

	return ds.db.Model(SettingModel{}).Where("id = ?", id).Update("value", string(bs)).Error
}

type SettingModel struct {
	ID        string    `gorm:"type:varchar(20);not null;primaryKey"`
	Value     string    `gorm:"type:varchar(1000);not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SettingModel) TableName() string {
	return "setting"
}
