package data

import (
	"testing"
)

func TestVacationAdd(t *testing.T) {
	v := NewVacationData()
	v.Add(Vacation{Destination: "Paris", StartDate: "2025-07-14", EndDate: "2025-07-18", Approved: true})
	if v.Len() != 1 {
		t.Errorf("expected 1, got %d", v.Len())
	}
}

func TestVacationRemove(t *testing.T) {
	v := NewVacationData()
	v.Add(Vacation{Destination: "Paris", StartDate: "2025-07-14", EndDate: "2025-07-18"})
	v.Add(Vacation{Destination: "London", StartDate: "2025-08-04", EndDate: "2025-08-08"})

	v.Remove("2025-07-14", "2025-07-18")
	if v.Len() != 1 {
		t.Errorf("expected 1, got %d", v.Len())
	}
	if v.All()[0].Destination != "London" {
		t.Error("wrong vacation remaining")
	}
}

func TestVacationRemoveNonExistent(t *testing.T) {
	v := NewVacationData()
	v.Add(Vacation{Destination: "Paris", StartDate: "2025-07-14", EndDate: "2025-07-18"})
	v.Remove("2025-01-01", "2025-01-05") // no match
	if v.Len() != 1 {
		t.Error("length should not change for no-match remove")
	}
}

func TestGetVacationMapWeekdays(t *testing.T) {
	v := NewVacationData()
	// Week of Mon 2025-01-06 to Fri 2025-01-10
	v.Add(Vacation{Destination: "Beach", StartDate: "2025-01-06", EndDate: "2025-01-10"})
	m := v.GetVacationMap()
	// Should have 5 weekdays
	if len(m) != 5 {
		t.Errorf("expected 5 weekdays, got %d", len(m))
	}
	if _, ok := m["2025-01-06"]; !ok {
		t.Error("Monday should be included")
	}
	if _, ok := m["2025-01-10"]; !ok {
		t.Error("Friday should be included")
	}
}

func TestGetVacationMapExcludesWeekends(t *testing.T) {
	v := NewVacationData()
	// Full week including weekend: Mon-Sun 2025-01-06 to 2025-01-12
	v.Add(Vacation{Destination: "Mountains", StartDate: "2025-01-06", EndDate: "2025-01-12"})
	m := v.GetVacationMap()
	// Mon-Fri only = 5 days
	if len(m) != 5 {
		t.Errorf("expected 5, got %d", len(m))
	}
	if _, ok := m["2025-01-11"]; ok {
		t.Error("Saturday should be excluded")
	}
	if _, ok := m["2025-01-12"]; ok {
		t.Error("Sunday should be excluded")
	}
}

func TestGetVacationMapMultipleVacations(t *testing.T) {
	v := NewVacationData()
	v.Add(Vacation{Destination: "A", StartDate: "2025-01-06", EndDate: "2025-01-07"}) // Mon-Tue
	v.Add(Vacation{Destination: "B", StartDate: "2025-01-13", EndDate: "2025-01-13"}) // Mon only
	m := v.GetVacationMap()
	if len(m) != 3 {
		t.Errorf("expected 3, got %d", len(m))
	}
}

func TestGetVacationMapInvalidDates(t *testing.T) {
	v := NewVacationData()
	v.Add(Vacation{Destination: "Bad", StartDate: "not-a-date", EndDate: "also-bad"})
	m := v.GetVacationMap()
	if len(m) != 0 {
		t.Error("invalid dates should produce empty map")
	}
}

func TestGetVacationMapSingleDay(t *testing.T) {
	v := NewVacationData()
	v.Add(Vacation{Destination: "Day Trip", StartDate: "2025-01-06", EndDate: "2025-01-06"})
	m := v.GetVacationMap()
	if len(m) != 1 {
		t.Errorf("expected 1, got %d", len(m))
	}
}

func TestVacationSaveLoad(t *testing.T) {
	dir := t.TempDir()
	v := NewVacationData()
	v.Add(Vacation{Destination: "Paris", StartDate: "2025-07-14", EndDate: "2025-07-18", Approved: true})
	v.Add(Vacation{Destination: "London", StartDate: "2025-08-04", EndDate: "2025-08-08", Approved: false})

	if err := v.SaveTo(dir); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadVacationDataFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Len() != 2 {
		t.Errorf("expected 2, got %d", loaded.Len())
	}
	all := loaded.All()
	if all[0].Destination != "Paris" {
		t.Errorf("expected Paris, got %s", all[0].Destination)
	}
	if all[1].Approved {
		t.Error("expected London to be unapproved")
	}
}

func TestVacationLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	v, err := LoadVacationDataFrom(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Len() != 0 {
		t.Error("expected empty vacations for missing file")
	}
}
