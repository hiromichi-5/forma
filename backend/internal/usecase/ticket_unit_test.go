package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIsValidTicketPriority(t *testing.T) {
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
			assert.Equal(t, tt.want, isValidTicketPriority(tt.input))
		})
	}
}

func TestUUIDPtrEqual(t *testing.T) {
	t.Parallel()

	id1 := uuid.New()
	id2 := uuid.New()
	id1Copy := id1

	tests := []struct {
		name string
		a    *uuid.UUID
		b    *uuid.UUID
		want bool
	}{
		{"両方nilならtrue", nil, nil, true},
		{"aだけnilならfalse", nil, &id1, false},
		{"bだけnilならfalse", &id1, nil, false},
		{"同じ値ならtrue", &id1, &id1Copy, true},
		{"異なる値ならfalse", &id1, &id2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, uuidPtrEqual(tt.a, tt.b))
		})
	}
}
