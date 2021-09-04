package quick

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemPubSub(t *testing.T) {
	logf := func(format string, args ...interface{}) {}
	ps := newMemPubSub(logf)

	topic := "topic1"
	payload := "payload1"
	var wg sync.WaitGroup
	wg.Add(2)

	ps.Subscribe(topic, func(s string) {
		assert.Equal(t, payload, s)
		wg.Done()
	})
	ps.Subscribe(topic, func(s string) {
		time.Sleep(time.Second)
		assert.Equal(t, payload, s)
		wg.Done()
	})
	ps.Publish(topic, payload)

	wg.Wait()
	ps.Close()
}
