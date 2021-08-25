package sms

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
)

type redisStorage struct {
	client *redis.Client
}

// NewRedisStorage 构造一个redis实现的Storage
func NewRedisStorage(client *redis.Client) Storage {
	return &redisStorage{
		client: client,
	}
}

// 以下方法实现 Storage

func (s *redisStorage) Get(key string) (*Code, bool) {
	var code Code
	err := s.client.Get(key).Scan(&code)
	if err != nil {
		return nil, false
	}
	return &code, true
}
func (s *redisStorage) Set(key string, code *Code) error {
	return s.client.Set(key, code, 60*time.Minute).Err()
}
func (s *redisStorage) Delete(key string) {
	if err := s.client.Del(key).Err(); err != nil {
		fmt.Println(err)
	}
}
