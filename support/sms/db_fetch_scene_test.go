package sms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDBFetchScene(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file=test.db?mode=memory"), &gorm.Config{})
	assert.Nil(t, err)

	db.AutoMigrate(SceneModel{})

	sm1 := &SceneModel{
		SceneID:           "LOGIN",
		Template:          "your code is {{code}}",
		ValidDuration:     300,
		ReproduceInterval: 60,
		CodeLen:           4,
	}
	assert.Nil(t, db.Create(sm1).Error)

	sm2 := &SceneModel{
		SceneID:           "REGISTER",
		Template:          "your code is {{code}}",
		ValidDuration:     300,
		ReproduceInterval: 60,
		CodeLen:           4,
	}
	assert.Nil(t, db.Create(sm2).Error)

	fetchScene := NewDBFetchScene(db)

	scene, err := fetchScene("LOGIN")
	assert.Nil(t, err)
	assert.Equal(t, sm1.Template, scene.Templdate)
	assert.Equal(t, sm1.ValidDuration, scene.ValidDuration)
	assert.Equal(t, sm1.ReproduceInterval, scene.ReproduceInterval)
	assert.Equal(t, sm1.CodeLen, scene.CodeLen)

	scene, err = fetchScene("REGISTER")
	assert.Nil(t, err)
	assert.Equal(t, sm2.Template, scene.Templdate)
	assert.Equal(t, sm2.ValidDuration, scene.ValidDuration)
	assert.Equal(t, sm2.ReproduceInterval, scene.ReproduceInterval)
	assert.Equal(t, sm2.CodeLen, scene.CodeLen)

	scene, err = fetchScene("NOTEXISTS")
	assert.Nil(t, scene)
	assert.NotNil(t, err)
}
