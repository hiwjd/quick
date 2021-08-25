package proxypool

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProxyPool(t *testing.T) {
	p1 := "http://127.0.0.1:9001"
	p2 := "http://127.0.0.1:9002"
	proxies := map[string]bool{
		p1: true,
		p2: true,
	}

	pool := New(Config{
		Logf: t.Logf,
		AliveCheck: func(u *url.URL) bool {
			return proxies[u.String()]
		},
		Verbose:       false,
		maxCheckTimes: 2,
	})
	// 一开始没有可用代理，候选代理也没有
	assert.Equal(t, -1, pool.idxLast)
	assert.Equal(t, 0, len(pool.candidates))
	t.Logf("%+v\n", pool.All())

	// 获取应该返回ErrNoValid
	_, err := pool.Get()
	assert.NotNil(t, err)
	assert.Equal(t, ErrNoValid, err)

	// 添加一个可用的代理
	err = pool.Add(p1)
	assert.Nil(t, err)
	assert.Equal(t, 0, pool.idxLast)
	assert.Equal(t, 0, len(pool.candidates))
	assert.Equal(t, 1, len(pool.All()))
	t.Logf("添加p1后的可用代理：%+v\n", pool.All())

	// 获取得到目前唯一可用的p1
	url, err := pool.Get()
	assert.Nil(t, err)
	assert.Equal(t, p1, url.String())

	// 10秒后 应该跑过一次存活检查了 Get结果应该跟前面一致
	time.Sleep(time.Second * 6)
	url, err = pool.Get()
	assert.Nil(t, err)
	assert.Equal(t, p1, url.String())

	// 再添加1个可用代理
	err = pool.Add(p2)
	assert.Nil(t, err)
	assert.Equal(t, 1, pool.idxLast)
	assert.Equal(t, 0, len(pool.candidates))
	assert.Equal(t, 2, len(pool.All()))
	t.Logf("添加p2后的可用代理：%+v\n", pool.All())

	// Get返回p1或p2中的1个
	url, err = pool.Get()
	assert.Nil(t, err)
	assert.Contains(t, []string{p1, p2}, url.String())

	// 关闭p1 等待一次存活检查，存活数1、候选数1
	t.Logf("关闭p1 %s\n", p1)
	proxies[p1] = false
	time.Sleep(time.Second * 6)

	assert.Equal(t, 0, pool.idxLast)
	assert.Equal(t, 1, len(pool.candidates))
	assert.Equal(t, 1, len(pool.All()))
	t.Logf("关闭p1后的可用代理：%+v\n", pool.All())

	// p1重新开启 等待存活检查后应回到存活
	t.Logf("开启p1 %s\n", p1)
	proxies[p1] = true
	time.Sleep(time.Second * 6)

	assert.Equal(t, 1, pool.idxLast)
	assert.Equal(t, 0, len(pool.candidates))
	assert.Equal(t, 2, len(pool.All()))
	t.Logf("重启p1后的可用代理：%+v\n", pool.All())

	// 关闭p2，执行3次存活检查后，p2从候选移除
	t.Logf("关闭p2，执行3次存活检查后，p2从候选移除")
	proxies[p2] = false
	time.Sleep(time.Second * 24)
	assert.Equal(t, 0, pool.idxLast)
	assert.Equal(t, 0, len(pool.candidates))
	assert.Equal(t, 1, len(pool.All()))
	t.Logf("关闭p2并执行3次存活检查后的可用代理：%+v\n", pool.All())

	pool.Stop()
}
