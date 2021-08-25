package sms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hiwjd/quick/util"
)

var (
	// ErrInvalid 验证码不存在或已失效
	ErrInvalid = errors.New("code is invalid")

	// ErrWrong 验证码错误
	ErrWrong = errors.New("code is wrong")
)

// ErrReproduceRemain 表明重发需要等待
type ErrReproduceRemain int64

// Error 实现error
func (e ErrReproduceRemain) Error() string {
	return fmt.Sprintf("需等待%d秒后重拾", e)
}

// Seconds 返回需要等待的秒数
func (e ErrReproduceRemain) Seconds() int64 {
	return int64(e)
}

// FnSkip 返回true时跳过检查
type FnSkip func(sceneID string, number string, code string) bool

// Validator 是验证码校验器
type Validator interface {

	// 生成验证码
	// sceneID 代表一个场景
	Produce(sceneID string, number string) (*Code, error)

	// 校验验证码
	// 成功校验后验证码即失效
	Validate(sceneID string, number string, code string) error
}

// NewValidator 构造Validator
func NewValidator(storage Storage, agency Agency, fetchScene FetchScene) Validator {
	return &validator{
		storage:    storage,
		agency:     agency,
		fetchScene: fetchScene,
		genCode:    defaultGenCode,
	}
}

// Code 是验证码
type Code struct {
	// 验证码
	code string
	// 生成时间
	produceAt int64
	// 有效时长 秒数
	validDuration int64
	// 再次生成的间隔 秒数
	reproduceInterval int64
}

// IsValid 验证码是否还有效
func (c Code) IsValid() bool {
	return c.produceAt+c.validDuration > time.Now().Unix()
}

// ReproduceRemain 可再次生产需要的秒数
func (c Code) ReproduceRemain() int64 {
	return c.produceAt + c.reproduceInterval - time.Now().Unix()
}

// MarshalBinary 实现encoding.BinaryMarshaler
func (c Code) MarshalBinary() (data []byte, err error) {
	m := map[string]interface{}{
		"code":              c.code,
		"produceAt":         c.produceAt,
		"validDuration":     c.validDuration,
		"reproduceInterval": c.reproduceInterval,
	}
	return json.Marshal(m)
}

// UnmarshalBinary 实现encoding.BinaryUnmarshaler
func (c *Code) UnmarshalBinary(data []byte) error {
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()

	m := make(map[string]interface{})
	if err := dec.Decode(&m); err != nil {
		return err
	}

	c.code = m["code"].(string)

	var err error
	if c.produceAt, err = m["produceAt"].(json.Number).Int64(); err != nil {
		return err
	}
	if c.validDuration, err = m["validDuration"].(json.Number).Int64(); err != nil {
		return err
	}
	if c.reproduceInterval, err = m["reproduceInterval"].(json.Number).Int64(); err != nil {
		return err
	}

	return nil
}

// Scene 是场景信息
type Scene struct {
	// 验证码内容模板
	Templdate string
	// 有效期 秒数
	ValidDuration int64
	// 可重发的间隔 秒数
	ReproduceInterval int64
	// 验证码位数
	CodeLen int
}

// FetchScene 提取场景信息
type FetchScene func(sceneID string) (*Scene, error)

// Storage 保存验证码数据
type Storage interface {
	// 获取 bool表示是否取到
	Get(key string) (*Code, bool)
	// 保存
	Set(key string, code *Code) error
	// 删除
	Delete(key string)
}

// GenCode 生成验证码
type GenCode func(codeLen int) string

var defaultGenCode = util.RandNumbers

// Skippable 可以设置跳过验证的方法
type Skippable interface {
	SetFnSkip(fnSkip FnSkip)
}

// GenCodeReplaceble 可以设置生成验证码方法
type GenCodeReplaceble interface {
	SetGenCode(genCode GenCode)
}

// validator 是默认的Validator实现
type validator struct {
	storage    Storage
	agency     Agency
	fetchScene FetchScene
	genCode    GenCode
	fnSkip     FnSkip
}

// Produce 实现Validator.Produce
func (v *validator) Produce(sceneID string, number string) (*Code, error) {
	key := fmt.Sprintf("%s-%s", sceneID, number)
	if code, ok := v.storage.Get(key); ok && code.IsValid() && code.ReproduceRemain() > 0 {
		// 存在且有效 且还没到重发的时间，返回错误：表示还没到重发的时间
		return code, ErrReproduceRemain(code.ReproduceRemain())
	}

	scene, err := v.fetchScene(sceneID)
	if err != nil {
		return nil, err
	}
	codeNum := v.genCode(scene.CodeLen)
	code := &Code{
		code:              codeNum,
		produceAt:         time.Now().Unix(),
		validDuration:     scene.ValidDuration,
		reproduceInterval: scene.ReproduceInterval,
	}

	if err := v.storage.Set(key, code); err != nil {
		return nil, err
	}

	content := util.Tpl(scene.Templdate, util.TplParam{"code": codeNum})
	if _, err := v.agency.Send(number, content); err != nil {
		v.storage.Delete(key)
		return nil, err
	}

	return code, nil
}

// Validate 实现Validator.Validate
func (v *validator) Validate(sceneID string, number string, cd string) error {
	if v.fnSkip != nil {
		if v.fnSkip(sceneID, number, cd) {
			return nil
		}
	}
	key := fmt.Sprintf("%s-%s", sceneID, number)

	var code *Code
	var ok bool
	if code, ok = v.storage.Get(key); !ok || !code.IsValid() {
		// 不存在 或者 无效
		return ErrInvalid
	}

	if code.code != cd {
		return ErrWrong
	}

	// 验证码成功后使该验证码失效
	v.storage.Delete(key)
	return nil
}

// SetFnSkip 设置跳过检查的方法
func (v *validator) SetFnSkip(fnSkip FnSkip) {
	v.fnSkip = fnSkip
}

// SetGenCode 设置验证码生成器
func (v *validator) SetGenCode(genCode GenCode) {
	v.genCode = genCode
}
