package cmd

import (
	"fmt"
	"io"
	"os"

	"rto/calc"
	"rto/data"
)

// RunStats prints statistics for the given period key to stdout.
func RunStats(periodKey string) error {
	td, err := data.LoadTimePeriodData()
	if err != nil {
		return fmt.Errorf("loading time periods: %w", err)
	}
	tp, err := td.GetPeriodByKey(periodKey)
	if err != nil {
		return fmt.Errorf("time period %q not found — run 'rto init' to create data files", periodKey)
	}

	badges, err := data.LoadBadgeEntryData()
	if err != nil {
		return fmt.Errorf("loading badge data: %w", err)
	}
	holidays, err := data.LoadHolidayData()
	if err != nil {
		return fmt.Errorf("loading holidays: %w", err)
	}
	vacations, err := data.LoadVacationData()
	if err != nil {
		return fmt.Errorf("loading vacations: %w", err)
	}

	settings, err := data.LoadAppSettings()
	if err != nil {
		return fmt.Errorf("loading settings: %w", err)
	}

	stats, err := calc.CalculatePeriodStats(tp, badges, holidays, vacations, settings.Goal, nil)
	if err != nil {
		return fmt.Errorf("calculating stats: %w", err)
	}

	return WriteStats(stats, os.Stdout)
}

// WriteStats formats and writes PeriodStats to the given writer.
func WriteStats(stats *calc.PeriodStats, w io.Writer) error {
	_, err := fmt.Fprintf(w, "Period: %s  (%s – %s)\n",
		stats.Name,
		stats.StartDate.Format("Jan 2, 2006"),
		stats.EndDate.Format("Jan 2, 2006"),
	)
	if err != nil {
		return err
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Status:               %s\n", stats.ComplianceStatus)
	fmt.Fprintf(w, "  Days ahead of pace:   %+d\n", stats.DaysAheadOfPace)
	if stats.RemainingMissableDays >= 0 {
		fmt.Fprintf(w, "  Skippable days left:  %d\n", stats.RemainingMissableDays)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Required badge-ins:   %d of %d total days (50%%)\n", stats.DaysRequired, stats.TotalDays)
	fmt.Fprintf(w, "  Badged in:            %d\n", stats.DaysBadgedIn)
	fmt.Fprintf(w, "  Still needed:         %d\n", stats.DaysStillNeeded)

	officeDays := stats.DaysBadgedIn - stats.FlexDays
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Badge-ins:            %d  (%d office, %d flex)\n", stats.DaysBadgedIn, officeDays, stats.FlexDays)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Days worked so far:   %d\n", stats.DaysThusFar)
	fmt.Fprintf(w, "  Days remaining:       %d\n", stats.DaysLeft)
	if stats.DaysThusFar > 0 {
		fmt.Fprintf(w, "  Current average:      %.1f%%\n", stats.CurrentAverage*100)
	}
	if stats.DaysLeft > 0 && stats.DaysStillNeeded > 0 {
		fmt.Fprintf(w, "  Rate needed:          %.1f%%\n", stats.RequiredFutureAverage*100)
	}

	if stats.ProjectedCompletionDate != nil {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  Projected completion: %s\n", stats.ProjectedCompletionDate.Format("Jan 2, 2006"))
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Holidays:             %d\n", stats.Holidays)
	fmt.Fprintf(w, "  Vacation days:        %d\n", stats.VacationDays)
	fmt.Fprintf(w, "  Days off (remote):    %d\n", stats.DaysOff)
	fmt.Fprintf(w, "  Available workdays:   %d\n", stats.AvailableWorkdays)

	return nil
}
