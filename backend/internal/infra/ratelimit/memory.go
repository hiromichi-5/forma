package ratelimit

import (
	"sync"
	"time"
)

type FixedWindow struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	now       func() time.Time
	counters  map[string]counter
	lastSweep time.Time
}

type counter struct {
	count int
	start time.Time
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:    limit,
		window:   window,
		now:      time.Now,
		counters: make(map[string]counter),
	}
}

func (l *FixedWindow) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if now.Sub(l.lastSweep) >= l.window {
		l.sweep(now)
	}

	c, ok := l.counters[key]
	if !ok || now.Sub(c.start) >= l.window {
		c = counter{start: now}
	}
	if c.count >= l.limit {
		return false
	}
	c.count++
	l.counters[key] = c
	return true
}

func (l *FixedWindow) sweep(now time.Time) {
	for k, c := range l.counters {
		if now.Sub(c.start) >= l.window {
			delete(l.counters, k)
		}
	}
	l.lastSweep = now
}
