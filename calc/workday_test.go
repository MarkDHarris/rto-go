package calc

import (
	"testing"
	"time"
)

func TestIsWeekday(t *testing.T) {
	tests := []struct {
		date    string
		want    bool
		weekday string
	}{
		{"2025-01-06", true, "Monday"},
		{"2025-01-07", true, "Tuesday"},
		{"2025-01-08", true, "Wednesday"},
		{"2025-01-09", true, "Thursday"},
		{"2025-01-10", true, "Friday"},
		{"2025-01-11", false, "Saturday"},
		{"2025-01-12", false, "Sunday"},
	}
	for _, tc := range tests {
		d, _ := time.Parse("2006-01-02", tc.date)
		got := IsWeekday(d)
		if got != tc.want {
			t.Errorf("%s (%s): expected %v, got %v", tc.date, tc.weekday, tc.want, got)
		}
	}
}

func TestCreateWorkdayMapCountsWeekdays(t *testing.T) {
	// Q1 2025: Jan 1 - Mar 31
	// Jan 1 is a Wednesday
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	m := CreateWorkdayMap(start, end)

	// January 2025 has 23 weekdays
	if len(m) != 23 {
		t.Errorf("expected 23 weekdays in January 2025, got %d", len(m))
	}
}

func TestCreateWorkdayMapExcludesWeekends(t *testing.T) {
	start := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC) // Monday
	end := time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)  // Sunday
	m := CreateWorkdayMap(start, end)

	if len(m) != 5 {
		t.Errorf("expected 5 weekdays (Mon-Fri), got %d", len(m))
	}
	if _, ok := m["2025-01-11"]; ok {
		t.Error("Saturday should not be in workday map")
	}
	if _, ok := m["2025-01-12"]; ok {
		t.Error("Sunday should not be in workday map")
	}
}

func TestCreateWorkdayMapIncludesBoundaries(t *testing.T) {
	start := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC) // Monday
	end := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)  // Friday
	m := CreateWorkdayMap(start, end)

	if _, ok := m["2025-01-06"]; !ok {
		t.Error("start date (Monday) should be included")
	}
	if _, ok := m["2025-01-10"]; !ok {
		t.Error("end date (Friday) should be included")
	}
	if len(m) != 5 {
		t.Errorf("expected 5, got %d", len(m))
	}
}

func TestCreateWorkdayMapWorkdayFlags(t *testing.T) {
	start := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	m := CreateWorkdayMap(start, end)

	wd, ok := m["2025-01-06"]
	if !ok {
		t.Fatal("expected entry for 2025-01-06")
	}
	if !wd.IsWorkday {
		t.Error("IsWorkday should be true")
	}
	if wd.IsBadgedIn {
		t.Error("IsBadgedIn should default to false")
	}
	if wd.IsHoliday {
		t.Error("IsHoliday should default to false")
	}
}

func TestCreateWorkdayMapSingleDay(t *testing.T) {
	// Single Monday
	d := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	m := CreateWorkdayMap(d, d)
	if len(m) != 1 {
		t.Errorf("expected 1, got %d", len(m))
	}

	// Single Saturday - should be empty
	sat := time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)
	m2 := CreateWorkdayMap(sat, sat)
	if len(m2) != 0 {
		t.Errorf("expected 0 for Saturday, got %d", len(m2))
	}
}

func TestCreateWorkdayMapKeyFormat(t *testing.T) {
	start := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	m := CreateWorkdayMap(start, end)

	if _, ok := m["2025-01-06"]; !ok {
		t.Error("key should be in YYYY-MM-DD format")
	}
}
