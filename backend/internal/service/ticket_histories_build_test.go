package service

import "testing"

func TestBuildTicketHistoryChanges_StatusAndPriority(t *testing.T) {
	changes := buildTicketHistoryChanges(
		true,
		"Open",
		"Closed",
		false,
		nil,
		nil,
		true,
		"medium",
		"high",
	)

	if len(changes) != 2 {
		t.Fatalf("want 2 changes, got %d", len(changes))
	}
	if changes[0].FieldName != "status" {
		t.Fatalf("want status change first, got %s", changes[0].FieldName)
	}
	if changes[0].OldValue == nil || *changes[0].OldValue != "Open" {
		t.Fatalf("unexpected status old value")
	}
	if changes[0].NewValue == nil || *changes[0].NewValue != "Closed" {
		t.Fatalf("unexpected status new value")
	}
	if changes[1].FieldName != "priority" {
		t.Fatalf("want priority change second, got %s", changes[1].FieldName)
	}
}

func TestBuildTicketHistoryChanges_AssigneeCleared(t *testing.T) {
	old := "Alice"
	changes := buildTicketHistoryChanges(
		false,
		"",
		"",
		true,
		&old,
		nil,
		false,
		"",
		"",
	)

	if len(changes) != 1 {
		t.Fatalf("want 1 change, got %d", len(changes))
	}
	if changes[0].FieldName != "assignee" {
		t.Fatalf("want assignee change, got %s", changes[0].FieldName)
	}
	if changes[0].OldValue == nil || *changes[0].OldValue != "Alice" {
		t.Fatalf("unexpected assignee old value")
	}
	if changes[0].NewValue != nil {
		t.Fatalf("new assignee should be nil")
	}
}

func TestBuildTicketHistoryChanges_NoChanges(t *testing.T) {
	changes := buildTicketHistoryChanges(false, "", "", false, nil, nil, false, "", "")
	if len(changes) != 0 {
		t.Fatalf("want no changes, got %d", len(changes))
	}
}
