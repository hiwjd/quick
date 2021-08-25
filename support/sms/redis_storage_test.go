package sms

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
)

func TestRedisStorage(t *testing.T) {
	// 需要启动redis
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	storage := NewRedisStorage(client)
	key := "some-key"

	code, ok := storage.Get(key)
	assert.False(t, ok)
	assert.Nil(t, code)

	code = &Code{
		code:              "1111",
		produceAt:         time.Now().Unix(),
		validDuration:     300,
		reproduceInterval: 60,
	}
	err := storage.Set(key, code)
	assert.Nil(t, err)

	code2, ok := storage.Get(key)
	assert.True(t, ok)
	assert.Equal(t, code, code2)

	// 清理
	storage.Delete(key)
}
