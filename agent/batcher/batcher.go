package batcher

import (
	"sync"
	"time"
)

type LogBatcher struct {
	mu        sync.Mutex
	buffer    []interface{}
	maxSize   int
	interval  time.Duration
	lastFlush time.Time
}

func New(max int, interval time.Duration) *LogBatcher {
	return &LogBatcher{
		maxSize:   max,
		interval:  interval,
		lastFlush: time.Now(),
	}
}

func (b *LogBatcher) Add(entry interface{}) (out []interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, entry)

	if len(b.buffer) >= b.maxSize || time.Since(b.lastFlush) >= b.interval {
		out = b.buffer
		b.buffer = nil
		b.lastFlush = time.Now()
	}
	return
}

func (b *LogBatcher) Flush() []interface{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buffer) == 0 {
		return nil
	}

	out := b.buffer
	b.buffer = nil
	b.lastFlush = time.Now()
	return out
}
