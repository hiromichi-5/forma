package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestTicketChangeStatus(t *testing.T) {
	t.Parallel()

	current := uuid.New()
	next := uuid.New()

	t.Run("異なるステータスなら変更前の値とtrueを返すこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{StatusID: current}
		old, changed := ticket.ChangeStatus(&FormStatus{ID: next})
		assert.True(t, changed)
		assert.Equal(t, current, old)
		assert.Equal(t, next, ticket.StatusID)
	})

	t.Run("同じステータスならfalseを返し値を変えないこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{StatusID: current}
		old, changed := ticket.ChangeStatus(&FormStatus{ID: current})
		assert.False(t, changed)
		assert.Equal(t, current, old)
		assert.Equal(t, current, ticket.StatusID)
	})

	t.Run("nilなら未指定として何も変えないこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{StatusID: current}
		old, changed := ticket.ChangeStatus(nil)
		assert.False(t, changed)
		assert.Equal(t, current, old)
		assert.Equal(t, current, ticket.StatusID)
	})
}

func TestTicketChangeAssignee(t *testing.T) {
	t.Parallel()

	current := uuid.New()
	next := uuid.New()

	t.Run("異なる担当者なら変更前の値とtrueを返すこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{AssigneeID: &current}
		old, changed := ticket.ChangeAssignee(SetAssignee(next))
		assert.True(t, changed)
		require.NotNil(t, old)
		assert.Equal(t, current, *old)
		require.NotNil(t, ticket.AssigneeID)
		assert.Equal(t, next, *ticket.AssigneeID)
	})

	t.Run("同じ担当者ならfalseを返すこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{AssigneeID: &current}
		_, changed := ticket.ChangeAssignee(SetAssignee(current))
		assert.False(t, changed)
	})

	t.Run("ClearAssigneeで担当者を解除できること", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{AssigneeID: &current}
		old, changed := ticket.ChangeAssignee(ClearAssignee())
		assert.True(t, changed)
		require.NotNil(t, old)
		assert.Nil(t, ticket.AssigneeID)
	})

	t.Run("未割当のままClearAssigneeしてもfalseを返すこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{}
		old, changed := ticket.ChangeAssignee(ClearAssignee())
		assert.False(t, changed)
		assert.Nil(t, old)
	})

	t.Run("KeepAssigneeなら未指定として何も変えないこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{AssigneeID: &current}
		old, changed := ticket.ChangeAssignee(KeepAssignee())
		assert.False(t, changed)
		require.NotNil(t, old)
		assert.Equal(t, &current, ticket.AssigneeID)
	})

	t.Run("ゼロ値のAssigneeChangeもKeepAssigneeと同じであること", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{AssigneeID: &current}
		_, changed := ticket.ChangeAssignee(AssigneeChange{})
		assert.False(t, changed)
	})
}

func TestTicketChangePriority(t *testing.T) {
	t.Parallel()

	high := PriorityHigh
	medium := PriorityMedium

	t.Run("異なる優先度なら変更前の値とtrueを返すこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{Priority: PriorityMedium}
		old, changed := ticket.ChangePriority(&high)
		assert.True(t, changed)
		assert.Equal(t, PriorityMedium, old)
		assert.Equal(t, PriorityHigh, ticket.Priority)
	})

	t.Run("同じ優先度ならfalseを返すこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{Priority: PriorityMedium}
		old, changed := ticket.ChangePriority(&medium)
		assert.False(t, changed)
		assert.Equal(t, PriorityMedium, old)
	})

	t.Run("nilなら未指定として何も変えないこと", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{Priority: PriorityMedium}
		old, changed := ticket.ChangePriority(nil)
		assert.False(t, changed)
		assert.Equal(t, PriorityMedium, old)
		assert.Equal(t, PriorityMedium, ticket.Priority)
	})
}
