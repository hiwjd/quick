package sms

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type dbSender struct {
	db *gorm.DB
}

// NewDBSender 构造DB实现的Sender
func NewDBSender(db *gorm.DB) Sender {
	return (&dbSender{db: db}).Send
}

// Send 实现Sender
func (s *dbSender) Send(number, content string) (RawResult, error) {
	q := &SendQueue{
		Number:  number,
		Content: content,
		Status:  0,
	}
	if err := s.db.Create(q).Error; err != nil {
		return nil, err
	}
	return RawResult{"message": "sms add to queue", "id": fmt.Sprintf("%d", q.ID)}, nil
}

// SendQueue 是短信发送队列
type SendQueue struct {
	ID        uint       `gorm:"primaryKey" json:"id" field:"id"`
	CreatedAt time.Time  `json:"createdAt" field:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" field:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt" field:"deleted_at"`
	Number    string     `gorm:"type:varchar(20)" json:"number" field:"number"`
	Content   string     `gorm:"type:varchar(255)" json:"content" field:"content"`
	// 发送状态 0:待发送 1:发送成功 2:发送失败
	Status byte `json:"status" field:"status"`
}

// TableName 定义表名
func (SendQueue) TableName() string {
	return "sms_queue"
}
