package sms

import (
	"errors"
	"net/http"
	"time"

	"github.com/hiwjd/quick"
	"gorm.io/gorm"
)

// NewDBFetchScene 返回数据库实现的FetchScene
func NewDBFetchScene(db *gorm.DB) FetchScene {
	return (&dbFetchScene{db}).Fetch
}

type dbFetchScene struct {
	db *gorm.DB
}

// SceneModel 是数据库中存放短信验证码场景模型
type SceneModel struct {
	// 场景ID
	SceneID string `gorm:"type:varchar(30);primay_key"`
	// 模板 格式：您好，您的验证码是{{code}} 当前预设的变量只有code
	Template string `gorm:"type:varchar(100);not null"`
	// 有效时长，单位秒
	ValidDuration int64 `gorm:"not null"`
	// 可重发的时间间隔，单位秒
	ReproduceInterval int64 `gorm:"not null"`
	// 验证码的位数
	CodeLen   int `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// TableName 指定表名
func (SceneModel) TableName() string {
	return "sms_scene"
}

func (s *dbFetchScene) Fetch(sceneID string) (*Scene, error) {
	var sm SceneModel
	if err := s.db.Where("scene_id=?", sceneID).First(&sm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = quick.NewFineErr(http.StatusInternalServerError, "短信场景缺失")
		}
		return nil, err
	}

	return &Scene{
		Templdate:         sm.Template,
		ValidDuration:     sm.ValidDuration,
		ReproduceInterval: sm.ReproduceInterval,
		CodeLen:           sm.CodeLen,
	}, nil
}
