package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"rto/calc"
	"rto/data"
)

func makeTestStats() *calc.PeriodStats {
	tp := &data.TimePeriod{
		Key:          "Q1_2025",
		Name:         "Q1",
		StartDateRaw: "2025-01-01",
		EndDateRaw:   "2025-03-31",
	}
	_ = tp.ParseDates()

	badges := data.NewBadgeEntryData()
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true})
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-03", IsBadgedIn: true, IsFlexCredit: true})
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-06", IsBadgedIn: true})

	holidays := data.NewHolidayData()
	holidays.Add(data.Holiday{Name: "New Year", Date: "2025-01-01"})

	vacations := data.NewVacationData()
	today := time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)

	stats, _ := calc.CalculatePeriodStats(tp, badges, holidays, vacations, 50, &today)
	return stats
}

func TestWriteStatsContainsPeriod(t *testing.T) {
	stats := makeTestStats()
	var buf bytes.Buffer
	if err := WriteStats(stats, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Q1") {
		t.Error("output should contain quarter name Q1")
	}
}

func TestWriteStatsContainsStatus(t *testing.T) {
	stats := makeTestStats()
	var buf bytes.Buffer
	WriteStats(stats, &buf)
	out := buf.String()
	if !strings.Contains(out, "Status:") {
		t.Error("output should contain Status field")
	}
}

func TestWriteStatsContainsBadgeCount(t *testing.T) {
	stats := makeTestStats()
	var buf bytes.Buffer
	WriteStats(stats, &buf)
	out := buf.String()
	if !strings.Contains(out, "3") {
		t.Error("output should contain badge count of 3")
	}
}

func TestWriteStatsContainsFlexCount(t *testing.T) {
	stats := makeTestStats()
	var buf bytes.Buffer
	WriteStats(stats, &buf)
	out := buf.String()
	if !strings.Contains(out, "flex") {
		t.Error("output should contain flex reference")
	}
}

func TestWriteStatsContainsHolidayCount(t *testing.T) {
	stats := makeTestStats()
	var buf bytes.Buffer
	WriteStats(stats, &buf)
	out := buf.String()
	if !strings.Contains(out, "Holidays:") {
		t.Error("output should contain Holidays field")
	}
}

func TestWriteStatsContainsVacationCount(t *testing.T) {
	stats := makeTestStats()
	var buf bytes.Buffer
	WriteStats(stats, &buf)
	out := buf.String()
	if !strings.Contains(out, "Vacation days:") {
		t.Error("output should contain Vacation days field")
	}
}

func TestWriteStatsContainsDatesRange(t *testing.T) {
	stats := makeTestStats()
	var buf bytes.Buffer
	WriteStats(stats, &buf)
	out := buf.String()
	if !strings.Contains(out, "2025") {
		t.Error("output should contain the year 2025")
	}
}

func TestWriteStatsContainsRequired(t *testing.T) {
	stats := makeTestStats()
	var buf bytes.Buffer
	WriteStats(stats, &buf)
	out := buf.String()
	if !strings.Contains(out, "Required badge-ins:") {
		t.Error("output should contain Required badge-ins field")
	}
}

func TestWriteStatsProjectedCompletion(t *testing.T) {
	tp := &data.TimePeriod{
		Key:          "Q1_2025",
		Name:         "Q1",
		StartDateRaw: "2025-01-01",
		EndDateRaw:   "2025-03-31",
	}
	_ = tp.ParseDates()

	badges := data.NewBadgeEntryData()
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true})

	today := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	stats, _ := calc.CalculatePeriodStats(tp, badges, data.NewHolidayData(), data.NewVacationData(), 50, &today)

	var buf bytes.Buffer
	WriteStats(stats, &buf)
	out := buf.String()

	if stats.ProjectedCompletionDate != nil && !strings.Contains(out, "Projected completion:") {
		t.Error("output should contain projected completion when set")
	}
}

func TestRunStatsWithTempDir(t *testing.T) {
	dir := t.TempDir()

	// Set up data: settings and time periods as separate files
	settings := data.AppSettings{DefaultOffice: "HQ", FlexCredit: "Flex", Goal: 50}
	if err := settings.SaveTo(dir); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	timePeriodFile := struct {
		TimePeriods []data.TimePeriod `yaml:"timeperiods"`
	}{
		TimePeriods: []data.TimePeriod{
			{Key: "Q1_2025", Name: "Q1", StartDateRaw: "2025-01-01", EndDateRaw: "2025-03-31"},
		},
	}
	if err := data.SaveYAMLTo(dir, "workday-fiscal-quarters.yaml", &timePeriodFile); err != nil {
		t.Fatalf("save time periods: %v", err)
	}
	badges := data.NewBadgeEntryData()
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true})
	if err := badges.SaveTo(dir); err != nil {
		t.Fatalf("save badges: %v", err)
	}
	hd := data.NewHolidayData()
	if err := hd.SaveTo(dir); err != nil {
		t.Fatalf("save holidays: %v", err)
	}
	vd := data.NewVacationData()
	if err := vd.SaveTo(dir); err != nil {
		t.Fatalf("save vacations: %v", err)
	}

	// Point global data dir at temp
	data.SetDataDir(dir)
	defer data.SetDataDir("")

	err := RunStats("Q1_2025")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRunStatsPeriodNotFound(t *testing.T) {
	dir := t.TempDir()
	settings := data.AppSettings{DefaultOffice: "HQ", FlexCredit: "Flex", Goal: 50}
	if err := settings.SaveTo(dir); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	timePeriodFile := struct {
		TimePeriods []data.TimePeriod `yaml:"timeperiods"`
	}{TimePeriods: []data.TimePeriod{}}
	if err := data.SaveYAMLTo(dir, "workday-fiscal-quarters.yaml", &timePeriodFile); err != nil {
		t.Fatalf("save time periods: %v", err)
	}

	data.SetDataDir(dir)
	defer data.SetDataDir("")

	err := RunStats("Q99_2025")
	if err == nil {
		t.Error("expected error for missing period")
	}
}

func TestWriteStatsNoCurrentAverage(t *testing.T) {
	// Stats with DaysThusFar = 0 should not show current average
	stats := &calc.PeriodStats{
		Name:             "Q1",
		StartDate:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
		ComplianceStatus: "Not Started",
		DaysThusFar:      0,
		DaysLeft:         63,
		DaysRequired:     32,
		DaysStillNeeded:  32,
		TotalDays:        63,
	}
	var buf bytes.Buffer
	WriteStats(stats, &buf)
	out := buf.String()
	if strings.Contains(out, "Current average") {
		t.Error("should not show current average when DaysThusFar is 0")
	}
}
