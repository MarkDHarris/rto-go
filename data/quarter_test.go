package data

import (
	"testing"
	"time"
)

func makeTimePeriodData(t *testing.T) *TimePeriodData {
	t.Helper()
	td := NewTimePeriodData()
	periods := []TimePeriod{
		{Key: "Q1_2025", Name: "Q1", StartDateRaw: "2025-01-01", EndDateRaw: "2025-03-31"},
		{Key: "Q2_2025", Name: "Q2", StartDateRaw: "2025-04-01", EndDateRaw: "2025-06-30"},
		{Key: "Q3_2025", Name: "Q3", StartDateRaw: "2025-07-01", EndDateRaw: "2025-09-30"},
		{Key: "Q4_2025", Name: "Q4", StartDateRaw: "2025-10-01", EndDateRaw: "2025-12-31"},
		{Key: "Q1_2026", Name: "Q1", StartDateRaw: "2026-01-01", EndDateRaw: "2026-03-31"},
	}
	for i := range periods {
		if err := periods[i].ParseDates(); err != nil {
			t.Fatalf("parse dates: %v", err)
		}
	}
	td.periods = periods
	return td
}

func TestPeriodGetByKey(t *testing.T) {
	td := makeTimePeriodData(t)

	p, err := td.GetPeriodByKey("Q2_2025")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "Q2" {
		t.Errorf("expected Q2, got %s", p.Name)
	}
}

func TestPeriodGetByKeyNotFound(t *testing.T) {
	td := makeTimePeriodData(t)
	_, err := td.GetPeriodByKey("Q4_2030")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestPeriodGetByDate(t *testing.T) {
	td := makeTimePeriodData(t)

	tests := []struct {
		date    string
		wantKey string
	}{
		{"2025-01-01", "Q1_2025"},
		{"2025-03-31", "Q1_2025"},
		{"2025-04-01", "Q2_2025"},
		{"2025-12-31", "Q4_2025"},
		{"2026-02-15", "Q1_2026"},
	}
	for _, tc := range tests {
		d, _ := time.Parse(BadgeDateFormat, tc.date)
		p, err := td.GetPeriodByDate(d)
		if err != nil {
			t.Errorf("date %s: unexpected error: %v", tc.date, err)
			continue
		}
		if p.Key != tc.wantKey {
			t.Errorf("date %s: expected %s, got %s", tc.date, tc.wantKey, p.Key)
		}
	}
}

func TestPeriodGetByDateOutOfRange(t *testing.T) {
	td := makeTimePeriodData(t)
	d, _ := time.Parse(BadgeDateFormat, "2030-06-01")
	_, err := td.GetPeriodByDate(d)
	if err == nil {
		t.Error("expected error for date outside all periods")
	}
}

func TestPeriodParseDates(t *testing.T) {
	p := TimePeriod{StartDateRaw: "2025-01-01", EndDateRaw: "2025-03-31", Key: "Q1_2025"}
	if err := p.ParseDates(); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if p.StartDate.Month() != time.January {
		t.Error("start month should be January")
	}
	if p.EndDate.Month() != time.March {
		t.Error("end month should be March")
	}
}

func TestPeriodParseDatesInvalid(t *testing.T) {
	p := TimePeriod{StartDateRaw: "not-a-date", EndDateRaw: "2025-03-31", Key: "Q1_2025"}
	if err := p.ParseDates(); err == nil {
		t.Error("expected error for invalid start date")
	}
}

func TestNearestPeriod(t *testing.T) {
	td := makeTimePeriodData(t)

	// Date within range
	d, _ := time.Parse(BadgeDateFormat, "2025-06-15")
	p, err := td.NearestPeriod(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Key != "Q2_2025" {
		t.Errorf("expected Q2_2025, got %s", p.Key)
	}

	// Date outside range (after all periods)
	d2, _ := time.Parse(BadgeDateFormat, "2027-01-01")
	p2, err := td.NearestPeriod(d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p2 == nil {
		t.Error("should return nearest period even when out of range")
	}
}

func TestTimePeriodKey(t *testing.T) {
	if TimePeriodKey("Q1", 2025) != "Q1_2025" {
		t.Errorf("expected Q1_2025")
	}
	if TimePeriodKey("q2", 2026) != "Q2_2026" {
		t.Errorf("expected Q2_2026")
	}
}

func TestPeriodSaveLoad(t *testing.T) {
	dir := t.TempDir()

	file := timePeriodDataFile{
		TimePeriods: []TimePeriod{
			{Key: "Q1_2025", Name: "Q1", StartDateRaw: "2025-01-01", EndDateRaw: "2025-03-31"},
		},
	}
	if err := SaveYAMLTo(dir, defaultTimePeriodsFilename, &file); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadTimePeriodDataFrom(dir, "")
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Len() != 1 {
		t.Errorf("expected 1, got %d", loaded.Len())
	}
	p := loaded.All()[0]
	if p.Key != "Q1_2025" {
		t.Errorf("expected Q1_2025, got %s", p.Key)
	}
	if p.StartDate.IsZero() {
		t.Error("StartDate should be parsed")
	}
}

func TestPeriodLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	td, err := LoadTimePeriodDataFrom(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if td.Len() != 0 {
		t.Error("expected empty periods for missing file")
	}
}

func TestPeriodAll(t *testing.T) {
	td := makeTimePeriodData(t)
	all := td.All()
	if len(all) != 5 {
		t.Errorf("expected 5, got %d", len(all))
	}
}

func TestParseDatesInvalidEnd(t *testing.T) {
	p := TimePeriod{StartDateRaw: "2025-01-01", EndDateRaw: "bad-date", Key: "Q1_2025"}
	if err := p.ParseDates(); err == nil {
		t.Error("expected error for invalid end date")
	}
}

func TestNearestPeriodEmpty(t *testing.T) {
	td := NewTimePeriodData()
	d, _ := time.Parse(BadgeDateFormat, "2025-06-15")
	_, err := td.NearestPeriod(d)
	if err == nil {
		t.Error("expected error for empty periods")
	}
}

func TestAbsDuration(t *testing.T) {
	if absDuration(-5*time.Second) != 5*time.Second {
		t.Error("expected positive duration for negative input")
	}
	if absDuration(3*time.Second) != 3*time.Second {
		t.Error("expected same duration for positive input")
	}
}

func TestGetCurrentPeriod(t *testing.T) {
	td := makeTimePeriodData(t)
	p, err := td.GetCurrentPeriod()
	if err == nil && p == nil {
		t.Error("if no error, period should not be nil")
	}
}

func TestCalendarDisplayColumnsDefault(t *testing.T) {
	dir := t.TempDir()
	file := timePeriodDataFile{
		TimePeriods: []TimePeriod{
			{Key: "Q1_2025", Name: "Q1", StartDateRaw: "2025-01-01", EndDateRaw: "2025-03-31"},
		},
	}
	if err := SaveYAMLTo(dir, defaultTimePeriodsFilename, &file); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := LoadTimePeriodDataFrom(dir, "")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.CalendarDisplayColumns() != 3 {
		t.Errorf("expected default CalendarDisplayColumns=3, got %d", loaded.CalendarDisplayColumns())
	}
}

func TestCalendarDisplayColumnsExplicit(t *testing.T) {
	dir := t.TempDir()
	file := timePeriodDataFile{
		CalendarDisplayColumns: 2,
		TimePeriods: []TimePeriod{
			{Key: "H1_2025", Name: "H1", StartDateRaw: "2025-01-01", EndDateRaw: "2025-06-30"},
		},
	}
	if err := SaveYAMLTo(dir, defaultTimePeriodsFilename, &file); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := LoadTimePeriodDataFrom(dir, "")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.CalendarDisplayColumns() != 2 {
		t.Errorf("expected CalendarDisplayColumns=2, got %d", loaded.CalendarDisplayColumns())
	}
}

func TestMonthSpan(t *testing.T) {
	tests := []struct {
		start, end string
		want       int
	}{
		{"2025-01-01", "2025-03-31", 3},
		{"2025-01-01", "2025-06-30", 6},
		{"2025-01-01", "2025-12-31", 12},
		{"2025-11-01", "2026-02-28", 4},
		{"2025-03-01", "2025-03-31", 1},
	}
	for _, tc := range tests {
		p := TimePeriod{StartDateRaw: tc.start, EndDateRaw: tc.end, Key: "test"}
		if err := p.ParseDates(); err != nil {
			t.Fatalf("parse: %v", err)
		}
		got := p.MonthSpan()
		if got != tc.want {
			t.Errorf("MonthSpan(%s, %s) = %d, want %d", tc.start, tc.end, got, tc.want)
		}
	}
}
