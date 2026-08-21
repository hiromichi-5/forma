package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriorityValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"high は有効", "high", true},
		{"medium は有効", "medium", true},
		{"low は有効", "low", true},
		{"空文字は無効", "", false},
		{"HIGH は無効", "HIGH", false},
		{"存在しない値 critical は無効", "critical", false},
		{"前後空白付きは無効", " high ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, Priority(tt.input).Valid())
		})
	}
}
