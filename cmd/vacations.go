package cmd

import (
	"fmt"
	"io"
	"os"

	"rto/data"
)

// RunVacations prints all vacations to stdout.
func RunVacations() error {
	vd, err := data.LoadVacationData()
	if err != nil {
		return fmt.Errorf("loading vacations: %w", err)
	}
	return WriteVacations(vd, os.Stdout)
}

// WriteVacations formats and writes vacation data to the given writer.
func WriteVacations(vd *data.VacationData, w io.Writer) error {
	all := vd.All()
	if len(all) == 0 {
		_, err := fmt.Fprintln(w, "No vacations recorded.")
		return err
	}

	// Header
	_, err := fmt.Fprintf(w, "%-4s  %-30s  %-12s  %-12s  %s\n",
		"#", "Destination", "Start", "End", "Approved")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%-4s  %-30s  %-12s  %-12s  %s\n",
		"----", "------------------------------", "------------", "------------", "--------")

	for i, v := range all {
		approved := "No"
		if v.Approved {
			approved = "Yes"
		}
		_, err := fmt.Fprintf(w, "%-4d  %-30s  %-12s  %-12s  %s\n",
			i+1, truncate(v.Destination, 30), v.StartDate, v.EndDate, approved)
		if err != nil {
			return err
		}
	}
	return nil
}

// truncate shortens a string to maxLen characters, adding "..." if needed.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
