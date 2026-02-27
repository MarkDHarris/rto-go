package calc

import (
	"time"
)

// Workday holds status flags for a single calendar day.
type Workday struct {
	Date         time.Time
	WorkDate     string // YYYY-MM-DD key
	IsWorkday    bool
	IsBadgedIn   bool
	IsFlexCredit bool
	IsHoliday    bool
	IsVacation   bool
}

// IsWeekday returns true if the date is Monday through Friday.
func IsWeekday(date time.Time) bool {
	w := date.Weekday()
	return w != time.Saturday && w != time.Sunday
}

// CreateWorkdayMap builds a map of YYYY-MM-DD → Workday for all weekdays in [start, end].
func CreateWorkdayMap(start, end time.Time) map[string]*Workday {
	m := make(map[string]*Workday)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if !IsWeekday(d) {
			continue
		}
		key := d.Format("2006-01-02")
		m[key] = &Workday{
			Date:      d,
			WorkDate:  key,
			IsWorkday: true,
		}
	}
	return m
}
