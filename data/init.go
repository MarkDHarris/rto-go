package data

import (
	"time"
)

// DefaultTimePeriods returns the standard set of quarterly time periods (Q1_2025 through Q4_2026).
func DefaultTimePeriods() []TimePeriod {
	type tpDef struct {
		key, name, start, end string
	}
	defs := []tpDef{
		{"Q1_2025", "Q1", "2025-01-01", "2025-03-31"},
		{"Q2_2025", "Q2", "2025-04-01", "2025-06-30"},
		{"Q3_2025", "Q3", "2025-07-01", "2025-09-30"},
		{"Q4_2025", "Q4", "2025-10-01", "2025-12-31"},
		{"Q1_2026", "Q1", "2026-01-01", "2026-03-31"},
		{"Q2_2026", "Q2", "2026-04-01", "2026-06-30"},
		{"Q3_2026", "Q3", "2026-07-01", "2026-09-30"},
		{"Q4_2026", "Q4", "2026-10-01", "2026-12-31"},
	}
	periods := make([]TimePeriod, len(defs))
	for i, d := range defs {
		periods[i] = TimePeriod{
			Key:          d.key,
			Name:         d.name,
			StartDateRaw: d.start,
			EndDateRaw:   d.end,
		}
	}
	return periods
}

// DefaultHolidays returns a standard set of US federal holidays for 2025–2026.
func DefaultHolidays() []Holiday {
	return []Holiday{
		{Name: "New Year's Day", Date: "2025-01-01"},
		{Name: "MLK Day", Date: "2025-01-20"},
		{Name: "Presidents' Day", Date: "2025-02-17"},
		{Name: "Memorial Day", Date: "2025-05-26"},
		{Name: "Juneteenth", Date: "2025-06-19"},
		{Name: "Independence Day", Date: "2025-07-04"},
		{Name: "Labor Day", Date: "2025-09-01"},
		{Name: "Columbus Day", Date: "2025-10-13"},
		{Name: "Veterans Day", Date: "2025-11-11"},
		{Name: "Thanksgiving Day", Date: "2025-11-27"},
		{Name: "Christmas Day", Date: "2025-12-25"},
		{Name: "New Year's Day", Date: "2026-01-01"},
		{Name: "MLK Day", Date: "2026-01-19"},
		{Name: "Presidents' Day", Date: "2026-02-16"},
		{Name: "Memorial Day", Date: "2026-05-25"},
		{Name: "Juneteenth", Date: "2026-06-19"},
		{Name: "Independence Day (observed)", Date: "2026-07-03"},
		{Name: "Labor Day", Date: "2026-09-07"},
		{Name: "Columbus Day", Date: "2026-10-12"},
		{Name: "Veterans Day", Date: "2026-11-11"},
		{Name: "Thanksgiving Day", Date: "2026-11-26"},
		{Name: "Christmas Day", Date: "2026-12-25"},
	}
}

// SampleBadgeEntry returns a sample badge entry using today's date.
func SampleBadgeEntry(office string) BadgeEntry {
	today := time.Now()
	return BadgeEntry{
		EntryDate:  today.Format(BadgeDateFormat),
		DateTime:   FlexTime{today},
		Office:     office,
		IsBadgedIn: true,
	}
}

// SampleVacation returns a sample vacation entry.
func SampleVacation() Vacation {
	return Vacation{
		Destination: "Vacation Destination",
		StartDate:   "2025-07-04",
		EndDate:     "2025-07-11",
		Approved:    true,
	}
}

// SampleEvent returns a sample event.
func SampleEvent() Event {
	return Event{
		Date:        time.Now().Format(BadgeDateFormat),
		Description: "Sample event",
	}
}
