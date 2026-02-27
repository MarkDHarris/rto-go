package data

import (
	"testing"
)

func TestHolidayAdd(t *testing.T) {
	h := NewHolidayData()
	h.Add(Holiday{Name: "New Year", Date: "2025-01-01"})
	h.Add(Holiday{Name: "MLK Day", Date: "2025-01-20"})
	if h.Len() != 2 {
		t.Errorf("expected 2, got %d", h.Len())
	}
}

func TestGetHolidayMap(t *testing.T) {
	h := NewHolidayData()
	h.Add(Holiday{Name: "New Year", Date: "2025-01-01"})
	h.Add(Holiday{Name: "MLK Day", Date: "2025-01-20"})
	h.Add(Holiday{Name: "Presidents Day", Date: "2025-02-17"})

	m := h.GetHolidayMap()
	if len(m) != 3 {
		t.Errorf("expected 3, got %d", len(m))
	}
	if m["2025-01-01"].Name != "New Year" {
		t.Error("wrong holiday name for New Year")
	}
	if _, ok := m["2025-01-02"]; ok {
		t.Error("2025-01-02 should not be in map")
	}
}

func TestHolidaySaveLoad(t *testing.T) {
	dir := t.TempDir()
	h := NewHolidayData()
	h.Add(Holiday{Name: "New Year", Date: "2025-01-01"})
	h.Add(Holiday{Name: "Memorial Day", Date: "2025-05-26"})

	if err := h.SaveTo(dir); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadHolidayDataFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Len() != 2 {
		t.Errorf("expected 2, got %d", loaded.Len())
	}
	m := loaded.GetHolidayMap()
	if m["2025-01-01"].Name != "New Year" {
		t.Error("holiday name mismatch")
	}
}

func TestHolidayLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	h, err := LoadHolidayDataFrom(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Len() != 0 {
		t.Error("expected empty holidays for missing file")
	}
}

func TestHolidayAll(t *testing.T) {
	h := NewHolidayData()
	h.Add(Holiday{Name: "New Year", Date: "2025-01-01"})
	all := h.All()
	if len(all) != 1 {
		t.Errorf("expected 1, got %d", len(all))
	}
	// Verify it's a copy
	all[0].Name = "Changed"
	all2 := h.All()
	if all2[0].Name == "Changed" {
		t.Error("All() should return a copy, not a reference")
	}
}

func TestHolidayDuplicateDate(t *testing.T) {
	h := NewHolidayData()
	h.Add(Holiday{Name: "Holiday A", Date: "2025-01-01"})
	h.Add(Holiday{Name: "Holiday B", Date: "2025-01-01"})
	// Last one wins in the map
	m := h.GetHolidayMap()
	if m["2025-01-01"].Name != "Holiday B" {
		t.Error("expected last holiday to win for duplicate date")
	}
}
