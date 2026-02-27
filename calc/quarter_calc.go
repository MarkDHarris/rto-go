package calc

import (
	"math"
	"time"

	"rto/data"
)

// PeriodStats contains all computed statistics for a single time period.
type PeriodStats struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time

	// Counts
	DaysBadgedIn      int
	FlexDays          int
	DaysThusFar       int
	DaysLeft          int
	TotalDays         int
	AvailableWorkdays int
	TotalCalendarDays int

	// Requirements
	DaysRequired          int
	DaysStillNeeded       int
	DaysOff               int
	Holidays              int
	VacationDays          int
	DaysAheadOfPace       int
	RemainingMissableDays int

	// Rates
	CurrentAverage        float64
	RequiredFutureAverage float64

	// Status
	ComplianceStatus string

	// Projection
	ProjectedCompletionDate *time.Time

	// Per-day status map
	WorkdayStats map[string]*Workday
}

// CalculatePeriodStats computes full statistics for a time period.
// goalPct is the required office percentage (e.g. 50 means 50%).
func CalculatePeriodStats(
	period *data.TimePeriod,
	badges *data.BadgeEntryData,
	holidays *data.HolidayData,
	vacations *data.VacationData,
	goalPct int,
	today *time.Time,
) (*PeriodStats, error) {
	now := time.Now()
	if today != nil {
		now = *today
	}
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	start := period.StartDate
	end := period.EndDate

	wdMap := CreateWorkdayMap(start, end)

	badgeMap := badges.GetBadgeMap(start, end)
	holidayMap := holidays.GetHolidayMap()
	vacationMap := vacations.GetVacationMap()

	totalCalendarDays := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		totalCalendarDays++
	}

	availableWorkdays := 0
	totalDays := 0
	daysBadgedIn := 0
	flexDays := 0
	daysThusFar := 0
	holidayCount := 0
	vacationDays := 0

	for dateKey, wd := range wdMap {
		if _, isHoliday := holidayMap[dateKey]; isHoliday {
			wd.IsHoliday = true
			holidayCount++
			availableWorkdays++
			continue
		}

		availableWorkdays++

		if _, isVacation := vacationMap[dateKey]; isVacation {
			wd.IsVacation = true
			vacationDays++
			continue
		}

		totalDays++

		if wd.Date.After(now) {
			if entry, ok := badgeMap[dateKey]; ok && entry.IsBadgedIn {
				wd.IsBadgedIn = true
				daysBadgedIn++
				if entry.IsFlexCredit {
					wd.IsFlexCredit = true
					flexDays++
				}
			}
			continue
		}

		isToday := wd.Date.Equal(now)
		if isToday {
			if entry, ok := badgeMap[dateKey]; ok && entry.IsBadgedIn {
				wd.IsBadgedIn = true
				daysBadgedIn++
				if entry.IsFlexCredit {
					wd.IsFlexCredit = true
					flexDays++
				}
			}
			wdMap[dateKey] = wd
			continue
		}

		daysThusFar++

		if entry, ok := badgeMap[dateKey]; ok && entry.IsBadgedIn {
			wd.IsBadgedIn = true
			daysBadgedIn++
			if entry.IsFlexCredit {
				wd.IsFlexCredit = true
				flexDays++
			}
		}
	}

	daysLeft := totalDays - daysThusFar
	daysRequired := int(math.Ceil(float64(totalDays) * float64(goalPct) / 100.0))
	daysStillNeeded := daysRequired - daysBadgedIn
	if daysStillNeeded < 0 {
		daysStillNeeded = 0
	}
	daysOff := daysThusFar - daysBadgedIn

	daysAheadOfPace := 0
	if daysThusFar > 0 && totalDays > 0 {
		expectedBadgeIns := int(math.Round(float64(daysThusFar) * float64(daysRequired) / float64(totalDays)))
		daysAheadOfPace = daysBadgedIn - expectedBadgeIns
	}

	remainingMissable := daysLeft - daysStillNeeded

	currentAverage := 0.0
	if daysThusFar > 0 {
		currentAverage = float64(daysBadgedIn) / float64(daysThusFar)
	}

	requiredFutureAverage := 0.0
	if daysLeft > 0 {
		requiredFutureAverage = float64(daysStillNeeded) / float64(daysLeft)
	}

	complianceStatus := determineComplianceStatus(daysBadgedIn, daysRequired, daysAheadOfPace, daysStillNeeded, daysLeft, end, now)

	var projectedDate *time.Time
	if daysBadgedIn > 0 && daysThusFar > 0 && daysStillNeeded > 0 {
		rate := float64(daysBadgedIn) / float64(daysThusFar)
		if rate > 0 {
			estimatedDays := int(math.Ceil(float64(daysStillNeeded) / rate))
			proj := now.AddDate(0, 0, estimatedDays)
			projectedDate = &proj
		}
	}

	return &PeriodStats{
		Name:                    period.Name,
		StartDate:               start,
		EndDate:                 end,
		DaysBadgedIn:            daysBadgedIn,
		FlexDays:                flexDays,
		DaysThusFar:             daysThusFar,
		DaysLeft:                daysLeft,
		TotalDays:               totalDays,
		AvailableWorkdays:       availableWorkdays,
		TotalCalendarDays:       totalCalendarDays,
		DaysRequired:            daysRequired,
		DaysStillNeeded:         daysStillNeeded,
		DaysOff:                 daysOff,
		Holidays:                holidayCount,
		VacationDays:            vacationDays,
		DaysAheadOfPace:         daysAheadOfPace,
		RemainingMissableDays:   remainingMissable,
		CurrentAverage:          currentAverage,
		RequiredFutureAverage:   requiredFutureAverage,
		ComplianceStatus:        complianceStatus,
		ProjectedCompletionDate: projectedDate,
		WorkdayStats:            wdMap,
	}, nil
}

func determineComplianceStatus(
	daysBadgedIn, daysRequired, daysAheadOfPace, daysStillNeeded, daysLeft int,
	periodEnd, today time.Time,
) string {
	if daysBadgedIn >= daysRequired {
		return "Achieved"
	}
	if daysAheadOfPace == 0 && daysBadgedIn == 0 {
		return "On Track"
	}
	if daysStillNeeded > daysLeft {
		return "Impossible"
	}
	if daysAheadOfPace < 0 {
		return "At Risk"
	}
	return "On Track"
}

// CalculateYearStats computes aggregate statistics across all time periods in a calendar year.
func CalculateYearStats(
	periods []*data.TimePeriod,
	badges *data.BadgeEntryData,
	holidays *data.HolidayData,
	vacations *data.VacationData,
	goalPct int,
	today *time.Time,
) (*PeriodStats, error) {
	if len(periods) == 0 {
		return nil, nil
	}

	start := periods[0].StartDate
	end := periods[0].EndDate
	for _, tp := range periods {
		if tp.StartDate.Before(start) {
			start = tp.StartDate
		}
		if tp.EndDate.After(end) {
			end = tp.EndDate
		}
	}

	syntheticTP := &data.TimePeriod{
		Key:          "Year",
		Name:         "Year",
		StartDate:    start,
		EndDate:      end,
		StartDateRaw: start.Format("2006-01-02"),
		EndDateRaw:   end.Format("2006-01-02"),
	}

	stats, err := CalculatePeriodStats(syntheticTP, badges, holidays, vacations, goalPct, today)
	if err != nil {
		return nil, err
	}
	stats.Name = "Year"
	return stats, nil
}
