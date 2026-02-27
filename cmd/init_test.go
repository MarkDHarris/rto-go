package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"rto/data"
)

func TestRunInitCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	if err := RunInitInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := []string{"settings.yaml", "workday-fiscal-quarters.yaml", "badge_data.json", "holidays.yaml", "vacations.yaml", "events.json"}
	for _, f := range files {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", f, err)
		}
	}
}

func TestRunInitCreatesDefaultTimePeriods(t *testing.T) {
	dir := t.TempDir()
	if err := RunInitInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tpd, err := data.LoadTimePeriodDataFrom(dir, "workday-fiscal-quarters.yaml")
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if tpd.Len() != 8 {
		t.Errorf("expected 8 periods (Q1_2025 through Q4_2026), got %d", tpd.Len())
	}

	// Check first and last
	tp, err := tpd.GetPeriodByKey("Q1_2025")
	if err != nil {
		t.Fatal("Q1_2025 should exist")
	}
	if tp.StartDateRaw != "2025-01-01" {
		t.Errorf("expected Q1_2025 start 2025-01-01, got %s", tp.StartDateRaw)
	}

	tp2, err := tpd.GetPeriodByKey("Q4_2026")
	if err != nil {
		t.Fatal("Q4_2026 should exist")
	}
	if tp2.EndDateRaw != "2026-12-31" {
		t.Errorf("expected Q4_2026 end 2026-12-31, got %s", tp2.EndDateRaw)
	}
}

func TestRunInitCreatesHolidays(t *testing.T) {
	dir := t.TempDir()
	if err := RunInitInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hd, err := data.LoadHolidayDataFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if hd.Len() == 0 {
		t.Error("expected some holidays to be created")
	}
}

func TestRunInitCreatesBadgeEntry(t *testing.T) {
	dir := t.TempDir()
	if err := RunInitInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bd, err := data.LoadBadgeEntryDataFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if bd.Len() == 0 {
		t.Error("expected at least one badge entry")
	}
}

func TestRunInitCreatesVacationEntry(t *testing.T) {
	dir := t.TempDir()
	if err := RunInitInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	vd, err := data.LoadVacationDataFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if vd.Len() == 0 {
		t.Error("expected at least one vacation entry")
	}
}

func TestRunInitCreatesEventEntry(t *testing.T) {
	dir := t.TempDir()
	if err := RunInitInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ed, err := data.LoadEventDataFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if ed.Len() == 0 {
		t.Error("expected at least one event entry")
	}
}

func TestRunInitCreatesDefaultSettings(t *testing.T) {
	dir := t.TempDir()
	if err := RunInitInDir(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, err := data.LoadAppSettingsFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if s.DefaultOffice == "" {
		t.Error("expected default office to be set")
	}
	if s.FlexCredit == "" {
		t.Error("expected flex credit label to be set")
	}
}

func TestSaveSettingsPreservesTimePeriods(t *testing.T) {
	dir := t.TempDir()
	if err := RunInitInDir(dir); err != nil {
		t.Fatalf("init error: %v", err)
	}

	// Get initial period count
	tpd, _ := data.LoadTimePeriodDataFrom(dir, "workday-fiscal-quarters.yaml")
	initialCount := tpd.Len()

	// Save new settings
	newSettings := data.AppSettings{DefaultOffice: "Denver, CO", FlexCredit: "WFH"}
	if err := newSettings.SaveTo(dir); err != nil {
		t.Fatalf("save settings error: %v", err)
	}

	// Verify periods preserved
	tpd2, err := data.LoadTimePeriodDataFrom(dir, "workday-fiscal-quarters.yaml")
	if err != nil {
		t.Fatalf("load periods error: %v", err)
	}
	if tpd2.Len() != initialCount {
		t.Errorf("expected %d periods preserved, got %d", initialCount, tpd2.Len())
	}

	// Verify settings updated
	s, _ := data.LoadAppSettingsFrom(dir)
	if s.DefaultOffice != "Denver, CO" {
		t.Errorf("expected Denver, CO, got %s", s.DefaultOffice)
	}
}
