package data

import (
	"testing"
	"time"
)

func TestBadgeEntryHas(t *testing.T) {
	b := NewBadgeEntryData()
	b.Add(BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true})
	if !b.Has("2025-01-02") {
		t.Error("expected Has to return true")
	}
	if b.Has("2025-01-03") {
		t.Error("expected Has to return false for missing entry")
	}
}

func TestBadgeEntryGet(t *testing.T) {
	b := NewBadgeEntryData()
	entry := BadgeEntry{EntryDate: "2025-01-06", Office: "HQ", IsBadgedIn: true}
	b.Add(entry)

	got, ok := b.Get("2025-01-06")
	if !ok {
		t.Fatal("expected to find entry")
	}
	if got.Office != "HQ" {
		t.Errorf("expected office HQ, got %s", got.Office)
	}
	_, ok = b.Get("2025-01-07")
	if ok {
		t.Error("expected not found for missing key")
	}
}

func TestBadgeEntryAdd(t *testing.T) {
	b := NewBadgeEntryData()
	if b.Len() != 0 {
		t.Error("expected empty")
	}
	b.Add(BadgeEntry{EntryDate: "2025-01-02"})
	b.Add(BadgeEntry{EntryDate: "2025-01-03"})
	if b.Len() != 2 {
		t.Errorf("expected 2, got %d", b.Len())
	}
}

func TestBadgeEntryRemove(t *testing.T) {
	b := NewBadgeEntryData()
	b.Add(BadgeEntry{EntryDate: "2025-01-02"})
	b.Add(BadgeEntry{EntryDate: "2025-01-03"})
	b.Add(BadgeEntry{EntryDate: "2025-01-06"})

	b.Remove("2025-01-03")
	if b.Len() != 2 {
		t.Errorf("expected 2, got %d", b.Len())
	}
	if b.Has("2025-01-03") {
		t.Error("removed entry should not exist")
	}
	if !b.Has("2025-01-02") {
		t.Error("other entries should remain")
	}
}

func TestBadgeEntryRemoveNonExistent(t *testing.T) {
	b := NewBadgeEntryData()
	b.Add(BadgeEntry{EntryDate: "2025-01-02"})
	b.Remove("2025-01-99") // should not panic
	if b.Len() != 1 {
		t.Error("length should not change")
	}
}

func TestGetBadgeMap(t *testing.T) {
	b := NewBadgeEntryData()
	b.Add(BadgeEntry{EntryDate: "2025-01-01", IsBadgedIn: true})
	b.Add(BadgeEntry{EntryDate: "2025-01-15", IsBadgedIn: true})
	b.Add(BadgeEntry{EntryDate: "2025-03-31", IsBadgedIn: true})

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	m := b.GetBadgeMap(start, end)

	if len(m) != 2 {
		t.Errorf("expected 2 entries in range, got %d", len(m))
	}
	if _, ok := m["2025-01-01"]; !ok {
		t.Error("expected 2025-01-01 in map")
	}
	if _, ok := m["2025-03-31"]; ok {
		t.Error("2025-03-31 should be outside range")
	}
}

func TestGetBadgeMapBoundaries(t *testing.T) {
	b := NewBadgeEntryData()
	b.Add(BadgeEntry{EntryDate: "2025-01-01", IsBadgedIn: true})
	b.Add(BadgeEntry{EntryDate: "2025-03-31", IsBadgedIn: true})

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)
	m := b.GetBadgeMap(start, end)

	if len(m) != 2 {
		t.Errorf("expected both boundary dates included, got %d entries", len(m))
	}
}

func TestBadgeEntryClone(t *testing.T) {
	b := NewBadgeEntryData()
	b.Add(BadgeEntry{EntryDate: "2025-01-02", Office: "HQ"})

	clone := b.Clone()
	clone.Remove("2025-01-02")

	if !b.Has("2025-01-02") {
		t.Error("original should not be affected by clone modification")
	}
}

func TestBadgeEntrySaveLoad(t *testing.T) {
	dir := t.TempDir()
	b := NewBadgeEntryData()
	b.Add(BadgeEntry{
		EntryDate:  "2025-01-02",
		DateTime:   FlexTime{time.Date(2025, 1, 2, 9, 0, 0, 0, time.UTC)},
		Office:     "McLean, VA",
		IsBadgedIn: true,
	})
	b.Add(BadgeEntry{
		EntryDate:    "2025-01-03",
		DateTime:     FlexTime{time.Date(2025, 1, 3, 8, 30, 0, 0, time.UTC)},
		Office:       "Flex Credit",
		IsBadgedIn:   true,
		IsFlexCredit: true,
	})

	if err := b.SaveTo(dir); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadBadgeEntryDataFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Len() != 2 {
		t.Errorf("expected 2 entries, got %d", loaded.Len())
	}
	e, _ := loaded.Get("2025-01-03")
	if !e.IsFlexCredit {
		t.Error("expected IsFlexCredit to be true")
	}
}

func TestBadgeEntryLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	b, err := LoadBadgeEntryDataFrom(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Len() != 0 {
		t.Error("expected empty badge data for missing file")
	}
}

func TestBadgeEntryAll(t *testing.T) {
	b := NewBadgeEntryData()
	b.Add(BadgeEntry{EntryDate: "2025-01-02", Office: "HQ"})
	b.Add(BadgeEntry{EntryDate: "2025-01-03", Office: "Remote"})

	all := b.All()
	if len(all) != 2 {
		t.Errorf("expected 2, got %d", len(all))
	}

	// Verify it's a copy — modifying returned slice shouldn't affect original
	all[0].Office = "CHANGED"
	orig, _ := b.Get("2025-01-02")
	if orig.Office == "CHANGED" {
		t.Error("All() should return a copy, not a reference")
	}
}

func TestBadgeEntryAllEmpty(t *testing.T) {
	b := NewBadgeEntryData()
	all := b.All()
	if len(all) != 0 {
		t.Errorf("expected 0, got %d", len(all))
	}
}
