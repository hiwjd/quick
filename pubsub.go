package quick

import (
	"sync"
)

type PubSub interface {
	// 发布事件
	Publish(topic string, payload string)
	// 订阅事件
	Subscribe(topic string, cb func(string))
	// 关闭
	Close()
}

func newMemPubSub(logf Logf) PubSub {
	return &memPubSub{
		done:        sync.WaitGroup{},
		mu:          sync.RWMutex{},
		subscribers: make(map[string][]chan string),
		logf:        logf,
	}
}

type memPubSub struct {
	done        sync.WaitGroup
	mu          sync.RWMutex
	subscribers map[string][]chan string
	logf        Logf
}

func (ps *memPubSub) Publish(topic string, payload string) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if cs, ok := ps.subscribers[topic]; ok {
		for _, c := range cs {
			c <- payload
		}
	}
}

func wf(logf Logf, cb func(string)) func(string) {
	return func(s string) {
		defer func() {
			if err := recover(); err != nil {
				logf("Subscriber Triggered Error: %#v", err)
			}
		}()
		cb(s)
	}
}

func (ps *memPubSub) Subscribe(topic string, cb func(string)) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	c := make(chan string, 8)
	if _, ok := ps.subscribers[topic]; ok {
		ps.subscribers[topic] = append(ps.subscribers[topic], c)
	} else {
		ps.subscribers[topic] = []chan string{c}
	}

	ps.done.Add(1)
	go func(_cb func(string)) {
		defer ps.done.Done()
		for payload := range c {
			_cb(payload)
		}
	}(wf(ps.logf, cb))
}

func (ps *memPubSub) Close() {
	for _, v := range ps.subscribers {
		for _, c := range v {
			close(c)
		}
	}
	ps.done.Wait()
}
