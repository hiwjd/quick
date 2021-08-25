package setting

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSetting(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file=test.db?mode=memory"), &gorm.Config{})
	assert.Nil(t, err)

	db.AutoMigrate(SettingModel{})

	st := New(db)

	ID := "CFD_PL_CONF"

	var data struct {
		Title    string
		ImageURL string
	}
	err = st.GetObj(ID, &data)
	assert.NotNil(t, err)
	assert.Empty(t, data.Title)
	assert.Empty(t, data.ImageURL)

	title := "测试标题"
	imageURL := "http://path/to/image"

	sm := SettingModel{
		ID:    ID,
		Value: fmt.Sprintf(`{"Title":"%s","ImageURL":"%s"}`, title, imageURL),
	}

	err = db.Create(sm).Error
	assert.Nil(t, err)

	err = st.GetObj(ID, &data)
	assert.Nil(t, err)
	assert.Equal(t, title, data.Title)
	assert.Equal(t, imageURL, data.ImageURL)
}
