package data

import (
	"testing"
)

func TestEventAdd(t *testing.T) {
	e := NewEventData()
	e.Add(Event{Date: "2025-01-15", Description: "Team meeting"})
	if e.Len() != 1 {
		t.Errorf("expected 1, got %d", e.Len())
	}
}

func TestEventAddSortsByDate(t *testing.T) {
	e := NewEventData()
	e.Add(Event{Date: "2025-03-01", Description: "C"})
	e.Add(Event{Date: "2025-01-01", Description: "A"})
	e.Add(Event{Date: "2025-02-01", Description: "B"})

	all := e.All()
	if all[0].Description != "A" {
		t.Errorf("expected A first, got %s", all[0].Description)
	}
	if all[1].Description != "B" {
		t.Errorf("expected B second, got %s", all[1].Description)
	}
	if all[2].Description != "C" {
		t.Errorf("expected C third, got %s", all[2].Description)
	}
}

func TestEventRemove(t *testing.T) {
	e := NewEventData()
	e.Add(Event{Date: "2025-01-15", Description: "Meeting"})
	e.Add(Event{Date: "2025-01-20", Description: "Lunch"})
	e.Add(Event{Date: "2025-01-15", Description: "Call"})

	e.Remove("2025-01-15", "Meeting")
	if e.Len() != 2 {
		t.Errorf("expected 2, got %d", e.Len())
	}
	// "Call" on same date should remain
	m := e.GetEventMap()
	calls := m["2025-01-15"]
	if len(calls) != 1 || calls[0].Description != "Call" {
		t.Error("Call event on 2025-01-15 should remain")
	}
}

func TestEventRemoveRequiresBothFields(t *testing.T) {
	e := NewEventData()
	e.Add(Event{Date: "2025-01-15", Description: "Meeting"})

	// Remove with wrong description - should not remove
	e.Remove("2025-01-15", "WrongDesc")
	if e.Len() != 1 {
		t.Error("remove with wrong description should not remove anything")
	}

	// Remove with wrong date - should not remove
	e.Remove("2025-01-16", "Meeting")
	if e.Len() != 1 {
		t.Error("remove with wrong date should not remove anything")
	}
}

func TestGetEventMap(t *testing.T) {
	e := NewEventData()
	e.Add(Event{Date: "2025-01-15", Description: "Morning standup"})
	e.Add(Event{Date: "2025-01-15", Description: "Lunch with team"})
	e.Add(Event{Date: "2025-01-20", Description: "All hands"})

	m := e.GetEventMap()
	if len(m["2025-01-15"]) != 2 {
		t.Errorf("expected 2 events on 2025-01-15, got %d", len(m["2025-01-15"]))
	}
	if len(m["2025-01-20"]) != 1 {
		t.Errorf("expected 1 event on 2025-01-20, got %d", len(m["2025-01-20"]))
	}
	if _, ok := m["2025-01-16"]; ok {
		t.Error("2025-01-16 should not be in map")
	}
}

func TestEventSaveLoad(t *testing.T) {
	dir := t.TempDir()
	e := NewEventData()
	e.Add(Event{Date: "2025-01-15", Description: "Team meeting"})
	e.Add(Event{Date: "2025-03-10", Description: "Performance review"})

	if err := e.SaveTo(dir); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadEventDataFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Len() != 2 {
		t.Errorf("expected 2, got %d", loaded.Len())
	}
}

func TestEventLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	e, err := LoadEventDataFrom(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Len() != 0 {
		t.Error("expected empty events for missing file")
	}
}

func TestEventAll(t *testing.T) {
	e := NewEventData()
	e.Add(Event{Date: "2025-01-15", Description: "Test"})
	all := e.All()
	all[0].Description = "Modified"
	// Original should not be modified
	all2 := e.All()
	if all2[0].Description == "Modified" {
		t.Error("All() should return a copy")
	}
}
