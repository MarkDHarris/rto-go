package app

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ── Style Helpers ────────────────────────────────────────────────────────────

func calendarDayStyle(selected, badged, flex, holidayVac, today, weekend, hasEvent bool) lipgloss.Style {
	style := lipgloss.NewStyle()

	switch {
	case selected:
		style = style.Background(lipgloss.Color("15")).Foreground(lipgloss.Color("0")).Bold(true)
		if today {
			style = style.Underline(true)
		}
	case badged && flex:
		style = style.Foreground(lipgloss.Color("172")).Bold(true)
	case badged:
		style = style.Foreground(lipgloss.Color("9")).Bold(true)
	case holidayVac:
		style = style.Foreground(lipgloss.Color("2"))
	case today:
		style = style.Bold(true).Underline(true)
	case weekend:
		style = style.Foreground(lipgloss.Color("240"))
	}

	if hasEvent && !selected {
		style = style.Foreground(lipgloss.Color("226"))
	}

	return style
}

func statusStyle(status string) lipgloss.Style {
	switch status {
	case "Achieved":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	case "On Track":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	case "At Risk":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	case "Impossible":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	default:
		return lipgloss.NewStyle()
	}
}

// ── Date Helpers ─────────────────────────────────────────────────────────────

// monthName returns the full English name for a month number (1–12).
func monthName(month int) string {
	names := []string{
		"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
	if month < 1 || month > 12 {
		return ""
	}
	return names[month]
}

// daysInMonth returns the number of days in a given month/year.
func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}

// addMonths adds n months to a date, clamping the day to the end of the month.
func addMonths(d time.Time, n int) time.Time {
	y := d.Year()
	m := int(d.Month()) + n
	for m > 12 {
		m -= 12
		y++
	}
	for m < 1 {
		m += 12
		y--
	}
	day := d.Day()
	maxDay := daysInMonth(y, m)
	if day > maxDay {
		day = maxDay
	}
	return time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.UTC)
}

// truncateStr shortens a string to maxLen runes (adds "..." if truncated).
func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
