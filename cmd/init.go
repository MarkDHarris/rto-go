package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"rto/data"
)

// RunInit initializes data files in the global data directory.
func RunInit() error {
	return RunInitInDir(data.GetDataDir())
}

// RunInitInDir initializes data files in the given directory (for testing).
// Existing files are never overwritten.
func RunInitInDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	settings := data.DefaultAppSettings()
	if !fileExists(dir, "settings.yaml") {
		if err := settings.SaveTo(dir); err != nil {
			return fmt.Errorf("writing settings.yaml: %w", err)
		}
	}

	tpFile := settings.ActiveTimePeriodFile(0)
	if !fileExists(dir, tpFile) {
		tpData := data.NewTimePeriodData()
		for _, tp := range data.DefaultTimePeriods() {
			tpData.Add(tp)
		}
		if err := tpData.SaveTo(dir); err != nil {
			return fmt.Errorf("writing %s: %w", tpFile, err)
		}
	}

	if !fileExists(dir, "badge_data.json") {
		badgeData := data.NewBadgeEntryData()
		badgeData.Add(data.SampleBadgeEntry(settings.DefaultOffice))
		if err := badgeData.SaveTo(dir); err != nil {
			return fmt.Errorf("writing badge_data.json: %w", err)
		}
	}

	if !fileExists(dir, "holidays.yaml") {
		holidayData := data.NewHolidayData()
		for _, h := range data.DefaultHolidays() {
			holidayData.Add(h)
		}
		if err := holidayData.SaveTo(dir); err != nil {
			return fmt.Errorf("writing holidays.yaml: %w", err)
		}
	}

	if !fileExists(dir, "vacations.yaml") {
		vacationData := data.NewVacationData()
		vacationData.Add(data.SampleVacation())
		if err := vacationData.SaveTo(dir); err != nil {
			return fmt.Errorf("writing vacations.yaml: %w", err)
		}
	}

	if !fileExists(dir, "events.json") {
		eventData := data.NewEventData()
		eventData.Add(data.SampleEvent())
		if err := eventData.SaveTo(dir); err != nil {
			return fmt.Errorf("writing events.json: %w", err)
		}
	}

	fmt.Printf("Initialized data files in: %s\n", dir)
	return nil
}

func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}
