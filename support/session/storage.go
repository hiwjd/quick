package session

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/hzxpz/ccc/util"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

var (
	// ErrSessionExpired 表示会话已经过期
	ErrSessionExpired = errors.New("会话已过期")
)

// Storage 是会话信息存储器
type Storage interface {
	Get(key string, session interface{}) error                          // 获取会话
	Set(session interface{}, ttl time.Duration) (key string, err error) // 存储会话 ttl=0表示不会失效
	RefreshTTL(key string, ttl time.Duration) error                     // 刷新会话存活时间 在原来过期时间基础上增加ttl
	Del(key string) error                                               // 删除会话
}

// RedisStorage redis实现的Storage
type RedisStorage struct {
	prefix string
	client *redis.Client
}

// NewRedisStorage 构造redis实现的Storage
func NewRedisStorage(prefix string, client *redis.Client) Storage {
	return &RedisStorage{
		prefix: prefix,
		client: client,
	}
}

func (s RedisStorage) unifyKey(key string) string {
	return s.prefix + key
}

// Get 实现Storage
func (s *RedisStorage) Get(key string, session interface{}) (err error) {
	var bs []byte
	if bs, err = s.client.Get(s.unifyKey(key)).Bytes(); err != nil {
		return
	}

	err = json.Unmarshal(bs, session)
	return
}

// Set 实现Storage
func (s *RedisStorage) Set(session interface{}, ttl time.Duration) (key string, err error) {
	var bs []byte
	if bs, err = json.Marshal(session); err != nil {
		return
	}
	key = xid.New().String()
	err = s.client.Set(s.unifyKey(key), bs, ttl).Err()
	return
}

// RefreshTTL 实现Storage
func (s *RedisStorage) RefreshTTL(key string, ttl time.Duration) error {
	return s.client.Expire(s.unifyKey(key), ttl).Err()
}

// Del 实现Storage
func (s *RedisStorage) Del(key string) error {
	return s.client.Del(s.unifyKey(key)).Err()
}

// AesStorage 是用aes加密用户信息的方式做会话存储的一种实现
type AesStorage struct {
	saes *util.SimpleAES
}

// NewAesStorage 构造AesStorage
func NewAesStorage(saes *util.SimpleAES) *AesStorage {
	return &AesStorage{
		saes: saes,
	}
}

// Get 实现Storage
func (s *AesStorage) Get(key string, session interface{}) error {
	bs, err := s.saes.Dec(key)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, session)
	return err
}

// Set 实现Storage
func (s *AesStorage) Set(session interface{}, ttl time.Duration) (key string, err error) {
	var bs []byte
	if bs, err = json.Marshal(session); err != nil {
		return
	}

	key, err = s.saes.Enc(bs)
	return
}

// RefreshTTL 实现Storage
func (s *AesStorage) RefreshTTL(key string, ttl time.Duration) error {
	return nil
}

// Del 实现Storage
func (s *AesStorage) Del(key string) error {
	return nil
}

// MysqlStorage 是存储在mysql中的会话
type MysqlStorage struct {
	db *gorm.DB
}

// NewMysqlStorage 构造MysqlStorage
func NewMysqlStorage(db *gorm.DB) *MysqlStorage {
	return &MysqlStorage{
		db: db,
	}
}

// Get 实现Storage
func (s *MysqlStorage) Get(key string, session interface{}) error {
	var ss Token
	if err := s.db.Where("token = ?", key).First(&ss).Error; err != nil {
		return err
	}

	if !ss.Permenent && ss.ExpireAt.Before(time.Now()) {
		return ErrSessionExpired
	}

	return json.Unmarshal([]byte(ss.Data), session)
}

// Set 实现Storage
func (s *MysqlStorage) Set(session interface{}, ttl time.Duration) (key string, err error) {
	var bs []byte
	if bs, err = json.Marshal(session); err != nil {
		return
	}

	key = xid.New().String()
	now := time.Now()
	ss := Token{
		Token:     key,
		Data:      string(bs),
		Permenent: ttl == 0,
		ExpireAt:  now.Add(ttl),
		CreatedAt: now,
	}

	err = s.db.Create(&ss).Error

	return
}

// RefreshTTL 实现Storage
func (s *MysqlStorage) RefreshTTL(key string, ttl time.Duration) error {
	var ss Token
	if err := s.db.Where("token=?", key).First(&ss).Error; err != nil {
		return err
	}

	ss.ExpireAt = ss.ExpireAt.Add(ttl)
	return s.db.Updates(ss).Error
}

// Del 实现Storage
func (s *MysqlStorage) Del(key string) error {
	return s.db.Where("token=?", key).Delete(Token{}).Error
}

// Token 是mysql中存储session的表
type Token struct {
	ID        uint   `gorm:"primaryKey"`
	Token     string `gorm:"type:varchar(30);uniqueIndex"`
	Data      string `gorm:"type:varchar(255)"`
	Permenent bool
	ExpireAt  time.Time
	CreatedAt time.Time
}
