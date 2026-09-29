package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFixedWindow_Allow(t *testing.T) {
	t.Parallel()

	newLimiter := func(now *time.Time) *FixedWindow {
		l := NewFixedWindow(2, time.Minute)
		l.now = func() time.Time { return *now }
		return l
	}

	t.Run("上限を超えた試行が拒否されること", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		l := newLimiter(&now)

		assert.True(t, l.Allow("a"))
		assert.True(t, l.Allow("a"))
		assert.False(t, l.Allow("a"))
	})

	t.Run("キーごとに独立して数えること", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		l := newLimiter(&now)

		assert.True(t, l.Allow("a"))
		assert.True(t, l.Allow("a"))
		assert.True(t, l.Allow("b"))
	})

	t.Run("制限が過ぎると再び許可されること", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		l := newLimiter(&now)

		assert.True(t, l.Allow("a"))
		assert.True(t, l.Allow("a"))
		assert.False(t, l.Allow("a"))

		now = now.Add(time.Minute)
		assert.True(t, l.Allow("a"))
	})
}
