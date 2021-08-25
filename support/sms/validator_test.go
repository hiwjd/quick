package sms

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	validatorBeTest Validator
	fnInSend        func(number, content string)
)

// mock Storage
type mockStorage struct {
	data map[string]*Code
}

func (ms *mockStorage) Get(key string) (*Code, bool) {
	if v, ok := ms.data[key]; ok {
		return v, ok
	}
	return nil, false
}

func (ms *mockStorage) Set(key string, code *Code) error {
	ms.data[key] = code
	return nil
}

func (ms *mockStorage) Delete(key string) {
	delete(ms.data, key)
}

func TestMain(m *testing.M) {
	// 准备环境

	// 短信发送中介
	agency, _ := NewAgency(func(number, content string) (RawResult, error) {
		// assert.Equal(t, tpl, content)
		if fnInSend != nil {
			fnInSend(number, content)
		}
		return RawResult{"message": "ok"}, nil
	})

	sceneMap := map[string]*Scene{
		"register": &Scene{
			Templdate:         "Your register code is {{code}}",
			ValidDuration:     300,
			ReproduceInterval: 6,
			CodeLen:           4,
		},
		"login": &Scene{
			Templdate:         "Your login code is {{code}}",
			ValidDuration:     300,
			ReproduceInterval: 6,
			CodeLen:           4,
		},
		"for-reproduce": &Scene{
			Templdate:         "message template",
			ValidDuration:     300,
			ReproduceInterval: 6,
			CodeLen:           4,
		},
		"for-validate": &Scene{
			Templdate:         "message template",
			ValidDuration:     3,
			ReproduceInterval: 3,
			CodeLen:           4,
		},
	}

	// 场景信息获取
	fetchScene := func(sceneID string) (*Scene, error) {
		return sceneMap[sceneID], nil
	}

	// 验证码存储器
	storage := &mockStorage{make(map[string]*Code)}

	// 测试的主角
	validatorBeTest = NewValidator(storage, agency, fetchScene)

	// 固定验证码 便于测试
	validatorBeTest.(*validator).genCode = func(codeLen int) string {
		return "1111"
	}

	os.Exit(m.Run())
}

// 测试校验
func TestValidatorValidateCase1(t *testing.T) {
	// 发送-成功校验-再次校验失效
	_, err := validatorBeTest.Produce("for-validate", "15812341234")
	assert.Nil(t, err)
	err = validatorBeTest.Validate("for-validate", "15812341234", "1111")
	assert.Nil(t, err)
	err = validatorBeTest.Validate("for-validate", "15812341234", "1111")
	assert.Equal(t, ErrInvalid, err)
}

func TestValidatorValidateCase2(t *testing.T) {
	// 发送-校验失败-再次校验成功-再次校验失效
	_, err := validatorBeTest.Produce("for-validate", "15812341235")
	assert.Nil(t, err)
	err = validatorBeTest.Validate("for-validate", "15812341235", "1112")
	assert.Equal(t, ErrWrong, err)
	err = validatorBeTest.Validate("for-validate", "15812341235", "1111")
	assert.Nil(t, err)
	err = validatorBeTest.Validate("for-validate", "15812341235", "1111")
	assert.Equal(t, ErrInvalid, err)
}

func TestValidatorValidateCase3(t *testing.T) {
	// 发送-过期去校验失效
	_, err := validatorBeTest.Produce("for-validate", "15812341236")
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)
	err = validatorBeTest.Validate("for-validate", "15812341236", "1111")
	assert.Equal(t, ErrInvalid, err)
}

// 测试场景信息获取及传入短信中介内中的内容
func TestValidatorScene(t *testing.T) {
	// 注册场景
	fnInSend = func(number, content string) {
		// 比对内容
		assert.Equal(t, "Your register code is 1111", content)
		assert.Equal(t, "15812341234", number)
	}
	_, err := validatorBeTest.Produce("register", "15812341234")
	assert.Nil(t, err)

	// 登录场景
	fnInSend = func(number, content string) {
		// 比对内容
		assert.Equal(t, "Your login code is 1111", content)
		assert.Equal(t, "15812341234", number)
	}
	_, err = validatorBeTest.Produce("login", "15812341234")
	assert.Nil(t, err)
}

// 测试重发控制
func TestValidatorReproduce(t *testing.T) {
	fnInSend = nil

	// 第一次发送
	code, err := validatorBeTest.Produce("for-reproduce", "15812341234")
	assert.Nil(t, err)
	assert.Equal(t, true, code.IsValid())
	assert.EqualValues(t, 6, code.ReproduceRemain())

	// 没到重发时间发送
	time.Sleep(4 * time.Second)
	code, err = validatorBeTest.Produce("for-reproduce", "15812341234")
	assert.IsType(t, ErrReproduceRemain(2), err)
	assert.Equal(t, true, code.IsValid())
	assert.True(t, code.ReproduceRemain() <= 2)

	// 过了重发时间发送
	time.Sleep(3 * time.Second)
	code, err = validatorBeTest.Produce("for-reproduce", "15812341234")
	assert.Nil(t, err)
	assert.Equal(t, true, code.IsValid())
	assert.EqualValues(t, 6, code.ReproduceRemain())
}

func TestCodeMarshalUnmarshalBinary(t *testing.T) {
	code := &Code{
		code:              "1111",
		produceAt:         1583063272,
		validDuration:     6,
		reproduceInterval: 3,
	}
	bs := []byte{
		123, 34, 99, 111, 100, 101, 34, 58, 34, 49, 49, 49, 49, 34, 44, 34,
		112, 114, 111, 100, 117, 99, 101, 65, 116, 34, 58, 49, 53, 56, 51,
		48, 54, 51, 50, 55, 50, 44, 34, 114, 101, 112, 114, 111, 100, 117,
		99, 101, 73, 110, 116, 101, 114, 118, 97, 108, 34, 58, 51, 44, 34,
		118, 97, 108, 105, 100, 68, 117, 114, 97, 116, 105, 111, 110, 34,
		58, 54, 125}

	bs2, err := code.MarshalBinary()
	assert.Nil(t, err)
	assert.Equal(t, bs, bs2)

	code2 := &Code{}
	err = code2.UnmarshalBinary(bs)
	assert.Nil(t, err)
	assert.Equal(t, code, code2)
}
