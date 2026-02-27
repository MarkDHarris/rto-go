package cmd

import (
	"fmt"
	"io"
	"os"

	"rto/data"
)

// RunHolidays prints all holidays to stdout.
func RunHolidays() error {
	hd, err := data.LoadHolidayData()
	if err != nil {
		return fmt.Errorf("loading holidays: %w", err)
	}
	return WriteHolidays(hd, os.Stdout)
}

// WriteHolidays formats and writes holiday data to the given writer.
func WriteHolidays(hd *data.HolidayData, w io.Writer) error {
	all := hd.All()
	if len(all) == 0 {
		_, err := fmt.Fprintln(w, "No holidays recorded.")
		return err
	}

	// Header
	_, err := fmt.Fprintf(w, "%-12s  %s\n", "Date", "Name")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%-12s  %s\n", "------------", "------------------------------")

	for _, h := range all {
		_, err := fmt.Fprintf(w, "%-12s  %s\n", h.Date, h.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
