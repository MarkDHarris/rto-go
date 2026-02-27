package cmd

import (
	"bytes"
	"strings"
	"testing"

	"rto/data"
)

func TestWriteVacationsEmpty(t *testing.T) {
	vd := data.NewVacationData()
	var buf bytes.Buffer
	if err := WriteVacations(vd, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No vacations") {
		t.Error("expected 'No vacations' message for empty list")
	}
}

func TestWriteVacationsShowsDestination(t *testing.T) {
	vd := data.NewVacationData()
	vd.Add(data.Vacation{Destination: "Paris", StartDate: "2025-07-14", EndDate: "2025-07-18", Approved: true})
	var buf bytes.Buffer
	WriteVacations(vd, &buf)
	out := buf.String()
	if !strings.Contains(out, "Paris") {
		t.Error("expected destination Paris in output")
	}
}

func TestWriteVacationsShowsDates(t *testing.T) {
	vd := data.NewVacationData()
	vd.Add(data.Vacation{Destination: "London", StartDate: "2025-08-04", EndDate: "2025-08-08"})
	var buf bytes.Buffer
	WriteVacations(vd, &buf)
	out := buf.String()
	if !strings.Contains(out, "2025-08-04") {
		t.Error("expected start date in output")
	}
	if !strings.Contains(out, "2025-08-08") {
		t.Error("expected end date in output")
	}
}

func TestWriteVacationsApprovedFlag(t *testing.T) {
	vd := data.NewVacationData()
	vd.Add(data.Vacation{Destination: "A", StartDate: "2025-01-01", EndDate: "2025-01-05", Approved: true})
	vd.Add(data.Vacation{Destination: "B", StartDate: "2025-02-01", EndDate: "2025-02-05", Approved: false})
	var buf bytes.Buffer
	WriteVacations(vd, &buf)
	out := buf.String()
	if !strings.Contains(out, "Yes") {
		t.Error("expected Yes for approved vacation")
	}
	if !strings.Contains(out, "No") {
		t.Error("expected No for unapproved vacation")
	}
}

func TestWriteVacationsHeader(t *testing.T) {
	vd := data.NewVacationData()
	vd.Add(data.Vacation{Destination: "Test", StartDate: "2025-01-01", EndDate: "2025-01-05"})
	var buf bytes.Buffer
	WriteVacations(vd, &buf)
	out := buf.String()
	if !strings.Contains(out, "Destination") {
		t.Error("expected Destination header")
	}
	if !strings.Contains(out, "Approved") {
		t.Error("expected Approved header")
	}
}

func TestTruncate(t *testing.T) {
	if truncate("short", 10) != "short" {
		t.Error("short string should not be truncated")
	}
	long := "This is a very long destination name"
	result := truncate(long, 20)
	if len([]rune(result)) > 20 {
		t.Errorf("truncated string too long: %d", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("truncated string should end with ...")
	}
}
