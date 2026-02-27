package cmd

import (
	"bytes"
	"strings"
	"testing"

	"rto/data"
)

func TestWriteHolidaysEmpty(t *testing.T) {
	hd := data.NewHolidayData()
	var buf bytes.Buffer
	if err := WriteHolidays(hd, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No holidays") {
		t.Error("expected 'No holidays' message for empty list")
	}
}

func TestWriteHolidaysShowsDate(t *testing.T) {
	hd := data.NewHolidayData()
	hd.Add(data.Holiday{Name: "New Year's Day", Date: "2025-01-01"})
	var buf bytes.Buffer
	WriteHolidays(hd, &buf)
	out := buf.String()
	if !strings.Contains(out, "2025-01-01") {
		t.Error("expected date in output")
	}
}

func TestWriteHolidaysShowsName(t *testing.T) {
	hd := data.NewHolidayData()
	hd.Add(data.Holiday{Name: "Independence Day", Date: "2025-07-04"})
	var buf bytes.Buffer
	WriteHolidays(hd, &buf)
	out := buf.String()
	if !strings.Contains(out, "Independence Day") {
		t.Error("expected holiday name in output")
	}
}

func TestWriteHolidaysHeader(t *testing.T) {
	hd := data.NewHolidayData()
	hd.Add(data.Holiday{Name: "MLK Day", Date: "2025-01-20"})
	var buf bytes.Buffer
	WriteHolidays(hd, &buf)
	out := buf.String()
	if !strings.Contains(out, "Date") {
		t.Error("expected Date header")
	}
	if !strings.Contains(out, "Name") {
		t.Error("expected Name header")
	}
}

func TestWriteHolidaysMultiple(t *testing.T) {
	hd := data.NewHolidayData()
	hd.Add(data.Holiday{Name: "New Year's Day", Date: "2025-01-01"})
	hd.Add(data.Holiday{Name: "MLK Day", Date: "2025-01-20"})
	hd.Add(data.Holiday{Name: "Presidents' Day", Date: "2025-02-17"})
	var buf bytes.Buffer
	WriteHolidays(hd, &buf)
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// 2 header lines + 3 holiday lines = 5
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines, got %d", len(lines))
	}
}
