package calc

import (
	"testing"
	"time"

	"rto/data"
)

// parseDate is a helper to parse YYYY-MM-DD in tests.
func parseDate(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

// makeQ1 returns a Q1_2025 period config for testing.
func makeQ1() *data.TimePeriod {
	q := &data.TimePeriod{
		Key:          "Q1_2025",
		Name:         "Q1",
		StartDateRaw: "2025-01-01",
		EndDateRaw:   "2025-03-31",
	}
	if err := q.ParseDates(); err != nil {
		panic(err)
	}
	return q
}

func emptyData() (*data.BadgeEntryData, *data.HolidayData, *data.VacationData) {
	return data.NewBadgeEntryData(), data.NewHolidayData(), data.NewVacationData()
}

func TestNoBadgesNoWork(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()
	today := parseDate("2025-01-01")

	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.DaysBadgedIn != 0 {
		t.Errorf("expected 0 badge-ins, got %d", stats.DaysBadgedIn)
	}
	if stats.DaysThusFar != 0 {
		t.Errorf("expected 0 days thus far, got %d", stats.DaysThusFar)
	}
}

func TestAchievedStatus(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	// Q1 2025 has 63 workdays total. Badge every day for first 10.
	today := parseDate("2025-01-14") // end of week 2
	for dateKey := parseDate("2025-01-02"); !dateKey.After(today); dateKey = dateKey.AddDate(0, 0, 1) {
		if IsWeekday(dateKey) {
			badges.Add(data.BadgeEntry{
				EntryDate:  dateKey.Format("2006-01-02"),
				IsBadgedIn: true,
			})
		}
	}
	// Force end of quarter to check achieved
	endOfQ := parseDate("2025-03-31")
	// Badge ~32 days out of 63 total (>50%)
	for dateKey := parseDate("2025-01-02"); !dateKey.After(parseDate("2025-02-28")); dateKey = dateKey.AddDate(0, 0, 1) {
		if IsWeekday(dateKey) && !badges.Has(dateKey.Format("2006-01-02")) {
			badges.Add(data.BadgeEntry{
				EntryDate:  dateKey.Format("2006-01-02"),
				IsBadgedIn: true,
			})
		}
	}
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &endOfQ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ComplianceStatus != "Achieved" {
		t.Errorf("expected Achieved, got %s (badged %d of %d required)", stats.ComplianceStatus, stats.DaysBadgedIn, stats.DaysRequired)
	}
}

func TestOnTrackStatus(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	// Jan 1 is a holiday (New Year's Day) — first real workdays are Jan 2 (Thu) and Jan 3 (Fri)
	// Badge both Mon and Thu/Fri (2/2 = 100% > 50% needed)
	holidays.Add(data.Holiday{Name: "New Year's Day", Date: "2025-01-01"})
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true})
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-03", IsBadgedIn: true})

	today := parseDate("2025-01-03")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ComplianceStatus != "On Track" {
		t.Errorf("expected On Track, got %s", stats.ComplianceStatus)
	}
	if stats.DaysAheadOfPace <= 0 {
		t.Errorf("expected ahead of pace, got %d", stats.DaysAheadOfPace)
	}
}

func TestAtRiskStatus(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	// Badge only 1 of first 10 days — far behind pace
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true})

	today := parseDate("2025-01-14")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ComplianceStatus != "At Risk" {
		t.Errorf("expected At Risk, got %s (badged %d of %d required, ahead=%d)", stats.ComplianceStatus, stats.DaysBadgedIn, stats.DaysRequired, stats.DaysAheadOfPace)
	}
}

func TestImpossibleStatus(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	// No badges at all at end of quarter
	today := parseDate("2025-03-31")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ComplianceStatus != "Impossible" {
		t.Errorf("expected Impossible, got %s", stats.ComplianceStatus)
	}
}

func TestFlexCreditCounting(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	badges.Add(data.BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true, IsFlexCredit: false, Office: "HQ"})
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-03", IsBadgedIn: true, IsFlexCredit: true, Office: "Flex"})
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-06", IsBadgedIn: true, IsFlexCredit: true, Office: "Flex"})

	today := parseDate("2025-01-06")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.DaysBadgedIn != 3 {
		t.Errorf("expected 3 badge-ins, got %d", stats.DaysBadgedIn)
	}
	if stats.FlexDays != 2 {
		t.Errorf("expected 2 flex days, got %d", stats.FlexDays)
	}
}

func TestHolidaysExcludedFromTotal(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	// Jan 1 is a holiday
	holidays.Add(data.Holiday{Name: "New Year", Date: "2025-01-01"})
	// Jan 20 is MLK Day
	holidays.Add(data.Holiday{Name: "MLK Day", Date: "2025-01-20"})

	today := parseDate("2025-01-31")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Holidays != 2 {
		t.Errorf("expected 2 holidays, got %d", stats.Holidays)
	}
	// Available workdays should be same as without holidays (holidays are still available workdays)
	// TotalDays should be availableWorkdays - holidayCount - vacationDays
	expectedTotal := stats.AvailableWorkdays - 2
	if stats.TotalDays != expectedTotal {
		t.Errorf("expected TotalDays = %d, got %d", expectedTotal, stats.TotalDays)
	}
}

func TestVacationDaysExcludedFromTotal(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	// 5-day vacation: Jan 6-10
	vacations.Add(data.Vacation{Destination: "Beach", StartDate: "2025-01-06", EndDate: "2025-01-10"})

	today := parseDate("2025-01-31")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.VacationDays != 5 {
		t.Errorf("expected 5 vacation days, got %d", stats.VacationDays)
	}
	expectedTotal := stats.AvailableWorkdays - stats.VacationDays
	if stats.TotalDays != expectedTotal {
		t.Errorf("expected TotalDays = %d, got %d", expectedTotal, stats.TotalDays)
	}
}

func TestDaysRequired50Percent(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	today := parseDate("2025-03-31")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// daysRequired = ceil(totalDays * 50 / 100)
	expected := (stats.TotalDays + 1) / 2
	if stats.DaysRequired != expected {
		t.Errorf("expected DaysRequired = %d, got %d", expected, stats.DaysRequired)
	}
}

func TestProjectedCompletionDateSet(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	// Badge 2 of first 4 days (50% rate, just enough)
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true})
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-06", IsBadgedIn: true})

	today := parseDate("2025-01-07")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.DaysStillNeeded > 0 && stats.ProjectedCompletionDate == nil {
		t.Error("expected projected completion date to be set when behind")
	}
}

func TestProjectedCompletionDateNilWhenAchieved(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	// Badge all days through end of quarter
	for d := parseDate("2025-01-02"); !d.After(parseDate("2025-03-31")); d = d.AddDate(0, 0, 1) {
		if IsWeekday(d) {
			badges.Add(data.BadgeEntry{EntryDate: d.Format("2006-01-02"), IsBadgedIn: true})
		}
	}

	today := parseDate("2025-03-31")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ComplianceStatus != "Achieved" {
		t.Fatalf("expected Achieved status")
	}
	if stats.ProjectedCompletionDate != nil {
		t.Error("projected completion date should be nil when already achieved (no more needed)")
	}
}

func TestDaysAheadOfPacePositive(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	// Badge every day - way ahead of pace
	for d := parseDate("2025-01-02"); !d.After(parseDate("2025-01-10")); d = d.AddDate(0, 0, 1) {
		if IsWeekday(d) {
			badges.Add(data.BadgeEntry{EntryDate: d.Format("2006-01-02"), IsBadgedIn: true})
		}
	}

	today := parseDate("2025-01-10")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.DaysAheadOfPace <= 0 {
		t.Errorf("expected positive days ahead of pace, got %d", stats.DaysAheadOfPace)
	}
}

func TestWorkdayStatsMap(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()
	holidays.Add(data.Holiday{Name: "New Year", Date: "2025-01-01"})
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true})
	vacations.Add(data.Vacation{Destination: "A", StartDate: "2025-01-06", EndDate: "2025-01-06"})

	today := parseDate("2025-01-07")
	stats, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Jan 1 (holiday)
	if wd, ok := stats.WorkdayStats["2025-01-01"]; !ok || !wd.IsHoliday {
		t.Error("2025-01-01 should be marked as holiday")
	}

	// Jan 2 (badged)
	if wd, ok := stats.WorkdayStats["2025-01-02"]; !ok || !wd.IsBadgedIn {
		t.Error("2025-01-02 should be marked as badged")
	}

	// Jan 6 (vacation)
	if wd, ok := stats.WorkdayStats["2025-01-06"]; !ok || !wd.IsVacation {
		t.Error("2025-01-06 should be marked as vacation")
	}

	// Weekend not in map
	if _, ok := stats.WorkdayStats["2025-01-04"]; ok {
		t.Error("Saturday 2025-01-04 should not be in workday stats")
	}
}

func TestCalculateYearStats(t *testing.T) {
	periods := []*data.TimePeriod{
		func() *data.TimePeriod {
			q := &data.TimePeriod{Key: "Q1_2025", Name: "Q1", StartDateRaw: "2025-01-01", EndDateRaw: "2025-03-31"}
			q.ParseDates()
			return q
		}(),
		func() *data.TimePeriod {
			q := &data.TimePeriod{Key: "Q2_2025", Name: "Q2", StartDateRaw: "2025-04-01", EndDateRaw: "2025-06-30"}
			q.ParseDates()
			return q
		}(),
		func() *data.TimePeriod {
			q := &data.TimePeriod{Key: "Q3_2025", Name: "Q3", StartDateRaw: "2025-07-01", EndDateRaw: "2025-09-30"}
			q.ParseDates()
			return q
		}(),
		func() *data.TimePeriod {
			q := &data.TimePeriod{Key: "Q4_2025", Name: "Q4", StartDateRaw: "2025-10-01", EndDateRaw: "2025-12-31"}
			q.ParseDates()
			return q
		}(),
	}

	badges, holidays, vacations := emptyData()
	today := parseDate("2025-06-30")

	stats, err := CalculateYearStats(periods, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.Name != "Year" {
		t.Errorf("expected Year, got %s", stats.Name)
	}
	// Year stats should span Jan 1 - Dec 31
	if stats.StartDate.Month() != time.January {
		t.Error("year stats should start in January")
	}
	if stats.EndDate.Month() != time.December {
		t.Error("year stats should end in December")
	}
}

func TestCalculateYearStatsEmpty(t *testing.T) {
	badges, holidays, vacations := emptyData()
	today := parseDate("2025-06-30")

	stats, err := CalculateYearStats(nil, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats != nil {
		t.Error("expected nil stats for empty quarters")
	}
}

func TestConfigurableGoalPercentage(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()
	today := parseDate("2025-03-31")

	stats50, err := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats75, err := CalculatePeriodStats(q, badges, holidays, vacations, 75, &today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats75.DaysRequired <= stats50.DaysRequired {
		t.Errorf("75%% goal (%d) should require more days than 50%% goal (%d)",
			stats75.DaysRequired, stats50.DaysRequired)
	}
}

func TestTodayBadgeDoesNotIncreaseDaysThusFar(t *testing.T) {
	q := makeQ1()
	badges, holidays, vacations := emptyData()

	badges.Add(data.BadgeEntry{EntryDate: "2025-01-02", IsBadgedIn: true})
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-03", IsBadgedIn: true})

	today := parseDate("2025-01-06")

	// Without badging today
	s1, _ := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)

	// Badge today
	badges.Add(data.BadgeEntry{EntryDate: "2025-01-06", IsBadgedIn: true})
	s2, _ := CalculatePeriodStats(q, badges, holidays, vacations, 50, &today)

	if s2.DaysThusFar != s1.DaysThusFar {
		t.Errorf("badging today should not change daysThusFar: before=%d after=%d",
			s1.DaysThusFar, s2.DaysThusFar)
	}
	if s2.DaysBadgedIn != s1.DaysBadgedIn+1 {
		t.Errorf("badging today should increase daysBadgedIn by 1: before=%d after=%d",
			s1.DaysBadgedIn, s2.DaysBadgedIn)
	}
}
