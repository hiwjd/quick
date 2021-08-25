package alarm

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/hiwjd/quick/util"
	"gorm.io/gorm"
)

// Alarm 是警报器，用于记录那些需要被及时关注的错误情况
type Alarm interface {
	Report(msg string)
}

// AlarmFunc 简化Alarm使用
type AlarmFunc func(msg string)

// Report 实现Alarm
func (h AlarmFunc) Report(msg string) {
	h(msg)
}

// NewRedisStoreAlarm 构造RedisStoreAlarm
func NewRedisStoreAlarm(client *redis.Client) *RedisStoreAlarm {
	return &RedisStoreAlarm{
		client: client,
	}
}

// RedisStoreAlarm 是使用redis作为存储的Alarm
type RedisStoreAlarm struct {
	client *redis.Client
}

// Report 实现Alarm
func (s *RedisStoreAlarm) Report(msg string) {
	if err := s.client.LPush("bernard-alarm-messages", msg).Err(); err != nil {
		log.Printf("[ERROR] RedisStoreAlarm.Report: %s %s\n", msg, err.Error())
	}
}

// MysqlStoreAlarm 是使用mysql作为存储的Alarm
type MysqlStoreAlarm struct {
	db *gorm.DB
}

// NewMysqlStoreAlarm 构造MysqlStoreAlarm
func NewMysqlStoreAlarm(db *gorm.DB) *MysqlStoreAlarm {
	return &MysqlStoreAlarm{db}
}

// Report 实现Alarm
func (s *MysqlStoreAlarm) Report(msg string) {
	file, line, fn := util.Caller(3)
	caller := fmt.Sprintf("f:%s l:%d fn:%s", file, line, fn)
	if len(caller) > 255 {
		caller = caller[:255]
	}
	if len(msg) > 255 {
		msg = msg[:255]
	}
	alog := &AlarmLog{
		Msg:    msg,
		Caller: caller,
	}
	if err := s.db.Create(alog).Error; err != nil {
		log.Printf("[ERROR] MysqlStoreAlarm.Report: %s %s %s", caller, msg, err.Error())
	}
}

// AlarmLog 记录系统中第一时间需要开发人员注意的信息
type AlarmLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Msg       string    `gorm:"type:varchar(255)" json:"msg"`
	Caller    string    `gorm:"type:varchar(255)" json:"caller"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (AlarmLog) TableName() string {
	return "alarm_log"
}
