package sms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDBSender(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file=test.db?mode=memory"), &gorm.Config{})
	assert.Nil(t, err)

	db.AutoMigrate(SendQueue{})

	send := NewDBSender(db)
	result, err := send("15612341234", "sms message")
	assert.Nil(t, err)
	assert.Equal(t, "1", result["id"])
}
