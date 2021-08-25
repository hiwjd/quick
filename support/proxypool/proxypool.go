package proxypool

import (
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hzxpz/ccc/support/log"
)

var (
	ErrURLInvalid = errors.New("URL Invalid")
	ErrNoValid    = errors.New("No Valid")
)

type (
	ProxyPool struct {
		mu            sync.RWMutex
		alives        []*url.URL  // 0-idxLast为当前可用的代理
		idxLast       int         // 存活的最大下标
		candidates    []candidate // 等待再次检测转为可用代理
		timer         *time.Timer // 检测当前可用和待检测可用性的定时器
		ac            AliveCheck  // 检查代理可用的方法
		maxCheckTimes int         // 检测不可用这么多次后移除
		logf          log.Logf
		verbose       bool
	}

	GetProxy func() (*url.URL, error)

	AliveCheck func(*url.URL) bool

	Config struct {
		Logf          log.Logf   // 日志方法
		AliveCheck    AliveCheck // 是检查代理存活的方法
		Verbose       bool       // true:打印详细的日志
		maxCheckTimes int        // 候选代理检测失败超过该次数后移除
	}

	candidate struct {
		v      *url.URL
		pTimes int
	}
)

func New(conf Config) *ProxyPool {
	logf := conf.Logf
	if logf == nil {
		logf = func(format string, args ...interface{}) {}
	}
	mct := conf.maxCheckTimes
	if mct < 1 {
		mct = 10
	}
	pool := &ProxyPool{
		alives:        make([]*url.URL, 32),
		idxLast:       -1,
		candidates:    make([]candidate, 0),
		logf:          logf,
		maxCheckTimes: mct,
		verbose:       conf.Verbose,
	}

	ac := conf.AliveCheck
	if ac == nil {
		ac = genDefaultAliveCheck(pool)
	}
	pool.ac = ac

	go pool.loopCheck()

	return pool
}

// Add 添加一个代理
// rawurl: http://xxxx:yy
func (p *ProxyPool) Add(rawurl string) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	if !p.isAlive(u) {
		return ErrURLInvalid
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.add(u)

	return nil
}

func (p *ProxyPool) Get() (*url.URL, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	n := p.idxLast + 1
	if n < 1 {
		return nil, ErrNoValid
	}
	return p.alives[rand.Intn(n)], nil
}

func (p *ProxyPool) All() []*url.URL {
	p.mu.RLock()
	defer p.mu.RUnlock()

	n := p.idxLast + 1
	if n < 1 {
		return nil
	}

	items := make([]*url.URL, n)
	for i := 0; i < n; i++ {
		items[i] = p.alives[i]
	}

	return items
}

func (p *ProxyPool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for p.timer != nil {
		if p.timer.Stop() {
			p.timer = nil
			break
		}
	}
}

func (p *ProxyPool) loopCheck() {
	p.timer = time.AfterFunc(time.Second*5, func() {
		if p.verbose {
			p.logf("开始检查代理存活情况")
		}
		defer p.loopCheck()

		var dead []int
		var alive []*url.URL
		var drop []int

		p.mu.RLock()
		for i := 0; i < p.idxLast+1; i++ {
			if !p.isAlive(p.alives[i]) {
				if p.verbose {
					p.logf("%s 等待加入候选\n", p.alives[i].String())
				}
				dead = append(dead, i)
			}
		}

		for i, itm := range p.candidates {
			if p.isAlive(itm.v) {
				if p.verbose {
					p.logf("%s 等待加入存活并从候选移除\n", itm.v.String())
				}
				alive = append(alive, itm.v)
				drop = append(drop, i)
			} else {
				if p.candidates[i].pTimes > p.maxCheckTimes {
					if p.verbose {
						p.logf("%s 等待从候选移除\n", itm.v.String())
					}
					drop = append(drop, i)
				} else {
					p.candidates[i].pTimes++
					if p.verbose {
						p.logf("%s 探测次数+1: %d\n", itm.v.String(), p.candidates[i].pTimes)
					}
				}
			}
		}
		p.mu.RUnlock()

		if len(dead) > 0 || len(alive) > 0 || len(drop) > 0 {
			if p.verbose {
				p.logf("查的失效代理：%+v\n", dead)
			}
			p.mu.Lock()
			defer p.mu.Unlock()

			for i := range dead {
				p.candidates = append(p.candidates, candidate{v: p.alives[i], pTimes: 0})
				if i != p.idxLast {
					p.alives[i] = p.alives[p.idxLast]
				}
				p.idxLast--
			}

			for _, v := range alive {
				p.add(v)
			}

			for i := range drop {
				if i != len(p.candidates) {
					p.candidates[i] = p.candidates[len(p.candidates)-1]
				}
				p.candidates = p.candidates[:len(p.candidates)-1]
			}
		}
		if p.verbose {
			p.logf("存活检查结束")
		}
	})
}

func (p *ProxyPool) add(v *url.URL) {
	if p.idxLast >= cap(p.alives) {
		newAlives := make([]*url.URL, cap(p.alives)*2)
		copy(newAlives, p.alives)
		p.alives = newAlives
	}
	p.idxLast++
	p.alives[p.idxLast] = v
}

func (p *ProxyPool) isAlive(u *url.URL) bool {
	return p.ac(u)
}

func genDefaultAliveCheck(p *ProxyPool) AliveCheck {
	client := &http.Client{
		Timeout: time.Second,
	}
	return func(u *url.URL) bool {
		if p.verbose {
			p.logf("开始检查 %s 是否存活\n", u.String())
		}
		s := time.Now()
		defer func() {
			d := time.Since(s)
			if p.verbose {
				p.logf("检查结束 耗时 %d 毫秒\n", d.Milliseconds())
			}
		}()

		rsp, err := client.Get(u.String())
		if err != nil {
			return false
		}
		return rsp.StatusCode == 200
	}
}
