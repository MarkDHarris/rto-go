package data

import (
	"testing"
)

func TestDefaultAppSettings(t *testing.T) {
	s := DefaultAppSettings()
	if s.DefaultOffice != "McLean, VA" {
		t.Errorf("expected McLean, VA, got %s", s.DefaultOffice)
	}
	if s.FlexCredit != "Flex Credit" {
		t.Errorf("expected Flex Credit, got %s", s.FlexCredit)
	}
	if s.Goal != 50 {
		t.Errorf("expected default goal 50, got %d", s.Goal)
	}
}

func TestAppSettingsSaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := AppSettings{DefaultOffice: "New York, NY", FlexCredit: "Remote Day"}

	if err := s.SaveTo(dir); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadAppSettingsFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.DefaultOffice != "New York, NY" {
		t.Errorf("expected New York, NY, got %s", loaded.DefaultOffice)
	}
	if loaded.FlexCredit != "Remote Day" {
		t.Errorf("expected Remote Day, got %s", loaded.FlexCredit)
	}
}

func TestAppSettingsLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	s, err := LoadAppSettingsFrom(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return defaults
	if s.DefaultOffice != settingsDefaultOffice {
		t.Errorf("expected default office, got %s", s.DefaultOffice)
	}
}

func TestAppSettingsPreservesTimePeriods(t *testing.T) {
	dir := t.TempDir()

	// Write initial time periods to default file
	initialPeriods := timePeriodDataFile{
		TimePeriods: []TimePeriod{
			{Key: "Q1_2025", Name: "Q1", StartDateRaw: "2025-01-01", EndDateRaw: "2025-03-31"},
		},
	}
	if err := SaveYAMLTo(dir, defaultTimePeriodsFilename, &initialPeriods); err != nil {
		t.Fatalf("save time periods error: %v", err)
	}

	// Save initial settings
	initialSettings := AppSettings{DefaultOffice: "HQ", FlexCredit: "Flex"}
	if err := SaveYAMLTo(dir, settingsFilename, &initialSettings); err != nil {
		t.Fatalf("save settings error: %v", err)
	}

	// Save new settings (should not overwrite time-periods.yaml)
	newSettings := AppSettings{DefaultOffice: "Remote", FlexCredit: "WFH"}
	if err := newSettings.SaveTo(dir); err != nil {
		t.Fatalf("save settings error: %v", err)
	}

	// Verify time periods are preserved
	td, err := LoadTimePeriodDataFrom(dir, "")
	if err != nil {
		t.Fatalf("load time periods error: %v", err)
	}
	if td.Len() != 1 {
		t.Errorf("expected 1 time period preserved, got %d", td.Len())
	}
	if td.All()[0].Key != "Q1_2025" {
		t.Error("time period key should be preserved")
	}

	// Verify settings updated
	loaded, err := LoadAppSettingsFrom(dir)
	if err != nil {
		t.Fatalf("load settings error: %v", err)
	}
	if loaded.DefaultOffice != "Remote" {
		t.Errorf("expected Remote, got %s", loaded.DefaultOffice)
	}
}

func TestTimePeriodsSavePreservesSettings(t *testing.T) {
	dir := t.TempDir()

	// Write initial settings to settings.yaml
	initialSettings := AppSettings{DefaultOffice: "HQ", FlexCredit: "Flex"}
	if err := SaveYAMLTo(dir, settingsFilename, &initialSettings); err != nil {
		t.Fatalf("save settings error: %v", err)
	}

	// Save new time periods (should not overwrite settings.yaml)
	td := NewTimePeriodData()
	td.periods = []TimePeriod{
		{Key: "Q2_2025", Name: "Q2", StartDateRaw: "2025-04-01", EndDateRaw: "2025-06-30"},
	}
	if err := td.SaveTo(dir); err != nil {
		t.Fatalf("save time periods error: %v", err)
	}

	// Verify settings preserved
	loaded, err := LoadAppSettingsFrom(dir)
	if err != nil {
		t.Fatalf("load settings error: %v", err)
	}
	if loaded.DefaultOffice != "HQ" {
		t.Errorf("expected HQ, got %s", loaded.DefaultOffice)
	}
}

func TestAppSettingsGoalSaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := AppSettings{DefaultOffice: "HQ", FlexCredit: "Flex", Goal: 60}

	if err := s.SaveTo(dir); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadAppSettingsFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Goal != 60 {
		t.Errorf("expected goal 60, got %d", loaded.Goal)
	}
}

func TestAppSettingsGoalDefaultWhenZero(t *testing.T) {
	dir := t.TempDir()
	// Save settings without goal (zero value)
	s := AppSettings{DefaultOffice: "HQ", FlexCredit: "Flex", Goal: 0}
	if err := s.SaveTo(dir); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadAppSettingsFrom(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	// Should fall back to default when loaded goal is 0
	if loaded.Goal != settingsDefaultGoal {
		t.Errorf("expected default goal %d when saved as 0, got %d", settingsDefaultGoal, loaded.Goal)
	}
}
