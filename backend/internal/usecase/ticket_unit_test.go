package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

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
