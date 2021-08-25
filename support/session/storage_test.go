package session

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/hzxpz/ccc/util"
	"github.com/stretchr/testify/assert"
)

type fakeSession struct {
	ID   uint
	Name string
}

func TestRedisStorage(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	storage := NewRedisStorage("", client)

	sess := &fakeSession{ID: 1, Name: "admin"}
	key, err := storage.Set(sess, 3*time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, key)

	var v fakeSession
	err = storage.Get(key, &v)
	assert.Nil(t, err)
	assert.Equal(t, sess, &v)

	time.Sleep(3 * time.Second)

	var v1 fakeSession
	err = storage.Get(key, &v1)
	assert.NotNil(t, err)
}

func TestAesStorage(t *testing.T) {
	saes, err := util.NewSimpleAES([]byte(`1234567812345678`))
	assert.Nil(t, err)
	storage := NewAesStorage(saes)

	sess := &fakeSession{ID: 1, Name: "admin"}
	key, err := storage.Set(sess, 3*time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, key)

	var v fakeSession
	err = storage.Get(key, &v)
	assert.Nil(t, err)
	assert.Equal(t, sess, &v)
}
