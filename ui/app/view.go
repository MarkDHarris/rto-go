package app

import (
	"fmt"
	"strings"
	"time"

	"rto/data"

	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) View() string {
	if m.termWidth == 0 {
		return "Initializing..."
	}

	mainContent := ""
	switch m.currentView {
	case ViewCalendar:
		mainContent = m.renderCalendar()
	case ViewVacations:
		mainContent = m.renderVacations()
	case ViewHolidays:
		mainContent = m.renderHolidays()
	case ViewSettings:
		mainContent = m.renderSettings()
	case ViewYearStats:
		mainContent = m.renderYearStats()
	default:
		mainContent = "Unknown view"
	}

	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		mainContent += "\n" + statusStyle.Render(m.statusMsg)
	}

	return mainContent
}

func (m *AppModel) renderCalendar() string {
	if m.termWidth < 20 || m.termHeight < 10 {
		return "Terminal too small"
	}

	// Build day maps
	holidayMap := m.holidayData.GetHolidayMap()
	vacationMap := m.vacationData.GetVacationMap()
	eventMap := m.eventData.GetEventMap()

	currentPeriod, _ := m.timePeriodData.GetPeriodByDate(m.navDate)
	var badgeMap map[string]data.BadgeEntry
	if currentPeriod != nil {
		badgeMap = m.badgeData.GetBadgeMap(currentPeriod.StartDate, currentPeriod.EndDate)
	} else {
		year := m.today.Year()
		badgeMap = m.badgeData.GetBadgeMap(
			time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC),
		)
	}

	var calendar strings.Builder

	if m.isWhatIf() {
		whatIfStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226")).Background(lipgloss.Color("52"))
		calendar.WriteString(whatIfStyle.Render(" ⚠ WHAT-IF MODE  (press w to exit, q to discard & quit) ") + "\n")
	}

	if currentPeriod != nil {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231"))
		title := fmt.Sprintf(" %s  [%s – %s]",
			currentPeriod.Key,
			currentPeriod.StartDate.Format("Jan 2, 2006"),
			currentPeriod.EndDate.Format("Jan 2, 2006"),
		)
		calendar.WriteString(titleStyle.Render(title) + "\n\n")
	}

	months := m.periodMonths(currentPeriod)
	cols := m.timePeriodData.CalendarDisplayColumns()

	for i := 0; i < len(months); i += cols {
		end := i + cols
		if end > len(months) {
			end = len(months)
		}
		var rowStrs []string
		for _, month := range months[i:end] {
			rowStrs = append(rowStrs, m.drawMonth(month, badgeMap, holidayMap, vacationMap, eventMap))
		}
		calendar.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowStrs...))
		if end < len(months) {
			calendar.WriteString("\n")
		}
	}

	eventsAndHelp := m.renderEventsAndHelp(eventMap)
	calendarAndEvents := lipgloss.JoinVertical(lipgloss.Left, calendar.String(), "\n", eventsAndHelp)

	periodTitle := "Period Stats"
	if currentPeriod != nil {
		periodTitle = fmt.Sprintf("Period Stats: %s", currentPeriod.Key)
	}
	periodBox := renderBoxWithTitle(periodTitle, m.renderStats(), statRowWidth)

	statsSection := periodBox
	yearContent := m.renderYearStatsContent()
	if yearContent != "" {
		yearTitle := fmt.Sprintf("Year Stats: %d", m.today.Year())
		if currentPeriod != nil {
			yearTitle = fmt.Sprintf("Year Stats: %d", currentPeriod.StartDate.Year())
		}
		yearBox := renderBoxWithTitle(yearTitle, yearContent, statRowWidth)
		statsSection = lipgloss.JoinVertical(lipgloss.Left, periodBox, yearBox)
	}

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, calendarAndEvents, "  ", statsSection)
	return mainContent
}

func (m *AppModel) periodMonths(period *data.TimePeriod) []time.Time {
	if period != nil {
		startMonth := time.Date(period.StartDate.Year(), period.StartDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		endMonth := time.Date(period.EndDate.Year(), period.EndDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		var months []time.Time
		for mo := startMonth; !mo.After(endMonth); mo = addMonths(mo, 1) {
			months = append(months, mo)
		}
		return months
	}
	return []time.Time{m.navDate, addMonths(m.navDate, 1), addMonths(m.navDate, 2)}
}

func (m *AppModel) drawMonth(
	month time.Time,
	badges map[string]data.BadgeEntry,
	holidayMap map[string]data.Holiday,
	vacations map[string]data.Vacation,
	events map[string][]data.Event,
) string {
	var b strings.Builder

	// Month header
	header := fmt.Sprintf("%s %d", monthName(int(month.Month())), month.Year())
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Width(24).Align(lipgloss.Center).Render(header) + "\n")

	// Day headers
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" Su Mo Tu We Th Fr Sa  ") + "\n")

	firstDay := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	startWeekday := int(firstDay.Weekday())
	daysInM := daysInMonth(month.Year(), int(month.Month()))

	var days []string
	// Pad with empty cells
	for i := 0; i < startWeekday; i++ {
		days = append(days, "  ")
	}

	for day := 1; day <= daysInM; day++ {
		date := time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, time.UTC)
		key := date.Format("2006-01-02")

		isSelected := date.Equal(m.selectedDate)
		isToday := date.Equal(m.today)
		isWeekend := date.Weekday() == time.Saturday || date.Weekday() == time.Sunday

		_, isBadged := badges[key]
		isFlexCredit := false
		if entry, ok := badges[key]; ok {
			isFlexCredit = entry.IsFlexCredit
		}
		_, isHoliday := holidayMap[key]
		_, isVacation := vacations[key]
		_, hasEvent := events[key]

		style := calendarDayStyle(isSelected, isBadged, isFlexCredit, isHoliday || isVacation, isToday, isWeekend, hasEvent)
		days = append(days, style.Render(fmt.Sprintf("%2d", day)))
	}

	// Create rows
	var rows []string
	for i := 0; i < len(days); i += 7 {
		end := i + 7
		if end > len(days) {
			end = len(days)
		}
		rows = append(rows, strings.Join(days[i:end], " "))
	}
	b.WriteString(strings.Join(rows, "\n"))

	return b.String()
}

func (m *AppModel) renderStats() string {
	s := m.activeStats
	if s == nil {
		return "No stats available"
	}

	var b strings.Builder

	sectionStyle := lipgloss.NewStyle().Bold(true)

	b.WriteString(sectionStyle.Render("STATUS") + "\n")
	b.WriteString(renderStatRow("  Status", statusStyle(s.ComplianceStatus).Render(s.ComplianceStatus), "") + "\n")

	paceStr := fmt.Sprintf("%+d days", s.DaysAheadOfPace)
	if s.DaysAheadOfPace > 0 {
		paceStr += " ahead"
	}
	b.WriteString(renderStatRow("  Days Ahead of Pace", paceStr, "") + "\n")
	skippableLabel := fmt.Sprintf("  Skippable Days (%d left - %d needed)", s.DaysLeft, s.DaysStillNeeded)
	b.WriteString(renderStatRow(skippableLabel, fmt.Sprintf("%d", s.RemainingMissableDays), "") + "\n")
	b.WriteString("\n")

	b.WriteString(sectionStyle.Render("PROGRESS") + "\n")
	b.WriteString(renderStatRow("  Total Days", fmt.Sprintf("%d", s.TotalCalendarDays), "") + "\n")
	b.WriteString(renderStatRow("  Total Working Days", fmt.Sprintf("%d", s.AvailableWorkdays-s.Holidays), "") + "\n")
	b.WriteString(renderStatRow("  Available Working Days", fmt.Sprintf("%d", s.TotalDays), "") + "\n")

	goalPct := ""
	if s.TotalDays > 0 {
		goalPct = fmt.Sprintf("%.1f%%", float64(s.DaysRequired)/float64(s.TotalDays)*100)
	}
	goalLabel := fmt.Sprintf("  Goal (%d%% Required)", m.settings.Goal)
	b.WriteString(renderStatRow(goalLabel, fmt.Sprintf("%d / %d", s.DaysRequired, s.TotalDays), goalPct) + "\n")
	officePct := ""
	if s.DaysRequired > 0 {
		officePct = fmt.Sprintf("%.1f%%", float64(s.DaysBadgedIn)/float64(s.DaysRequired)*100)
	}
	b.WriteString(renderStatRow("  Office Days", fmt.Sprintf("%d / %d", s.DaysBadgedIn, s.DaysRequired), officePct) + "\n")
	badgeOnly := s.DaysBadgedIn - s.FlexDays
	badgePct := ""
	flexPct := ""
	if s.DaysBadgedIn > 0 {
		badgePct = fmt.Sprintf("%.1f%%", float64(badgeOnly)/float64(s.DaysBadgedIn)*100)
		flexPct = fmt.Sprintf("%.1f%%", float64(s.FlexDays)/float64(s.DaysBadgedIn)*100)
	}
	b.WriteString(renderStatRow("   Badge-In Days", fmt.Sprintf("%d", badgeOnly), badgePct) + "\n")
	b.WriteString(renderStatRow("   Flex Credits", fmt.Sprintf("%d", s.FlexDays), flexPct) + "\n")
	neededPct := ""
	if s.DaysRequired > 0 {
		neededPct = fmt.Sprintf("%.1f%%", float64(s.DaysStillNeeded)/float64(s.DaysRequired)*100)
	}
	b.WriteString(renderStatRow("  Still Needed", fmt.Sprintf("%d / %d", s.DaysStillNeeded, s.DaysRequired), neededPct) + "\n")

	return b.String()
}

const (
	statRowWidth = 62
	statPctCol   = 8
	statValueCol = 14
	statLabelCol = statRowWidth - statValueCol - statPctCol // 40
)

func renderStatRow(label, value, pct string) string {
	labelVisW := lipgloss.Width(label)
	labelPad := statLabelCol - labelVisW
	if labelPad < 1 {
		labelPad = 1
	}

	valueVisW := lipgloss.Width(value)
	valuePad := statValueCol - valueVisW
	if valuePad < 0 {
		valuePad = 0
	}

	var pctCol string
	if pct != "" {
		pctVisW := lipgloss.Width(pct)
		pctPad := statPctCol - pctVisW
		if pctPad < 0 {
			pctPad = 0
		}
		pctCol = strings.Repeat(" ", pctPad) + pct
	} else {
		pctCol = strings.Repeat(" ", statPctCol)
	}

	return label + strings.Repeat(" ", labelPad) + strings.Repeat(" ", valuePad) + value + pctCol
}

func renderBoxWithTitle(title, content string, minWidth ...int) string {
	lines := strings.Split(content, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	contentW := 0
	for _, l := range lines {
		w := lipgloss.Width(l)
		if w > contentW {
			contentW = w
		}
	}

	titlePartLen := len(title) + 2
	if contentW < titlePartLen {
		contentW = titlePartLen
	}
	if len(minWidth) > 0 && contentW < minWidth[0] {
		contentW = minWidth[0]
	}

	borderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231"))

	var b strings.Builder

	dashCount := contentW - titlePartLen
	if dashCount < 0 {
		dashCount = 0
	}
	b.WriteString(borderStyle.Render("┌ ") + titleStyle.Render(title) + borderStyle.Render(" "+strings.Repeat("─", dashCount)+"┐") + "\n")

	for _, l := range lines {
		w := lipgloss.Width(l)
		pad := contentW - w
		if pad < 0 {
			pad = 0
		}
		b.WriteString(borderStyle.Render("│") + l + strings.Repeat(" ", pad) + borderStyle.Render("│") + "\n")
	}

	b.WriteString(borderStyle.Render("└" + strings.Repeat("─", contentW) + "┘"))
	return b.String()
}

func (m *AppModel) renderEventsAndHelp(eventMap map[string][]data.Event) string {
	var b strings.Builder
	dateKey := m.selectedDate.Format("2006-01-02")
	dayEvents := eventMap[dateKey]

	eventStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	b.WriteString(eventStyle.Bold(true).Render(fmt.Sprintf(" Events for %s:", m.selectedDate.Format("Mon Jan 2, 2006"))) + "\n")

	if len(dayEvents) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  (none)") + "\n")
	} else {
		for _, ev := range dayEvents {
			b.WriteString("  • " + ev.Description + "\n")
		}
	}

	if m.mode == ModeAdd {
		b.WriteString("\n" + eventStyle.Render(" Add event: "+m.inputBuffer+"_") + "\n")
	} else if m.mode == ModeDelete {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(" Delete event (type description): "+m.inputBuffer+"_") + "\n")
	} else if m.mode == ModeSearch {
		b.WriteString("\n" + eventStyle.Render(" Search: "+m.inputBuffer+"_") + "\n")
		// if len(m.searchResults) > 0 {
		// 	for _, ev := range m.searchResults {
		// 		b.WriteString(fmt.Sprintf("  [%s] %s\n", ev.Date, ev.Description))
		// 	}
		// }
	}

	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))

	viewLabel := m.settings.ActiveTimePeriodFile(m.activeTimePeriodIdx)
	if len(m.settings.TimePeriods) > 1 {
		viewLabel += fmt.Sprintf("  (%d of %d)", m.activeTimePeriodIdx+1, len(m.settings.TimePeriods))
	}
	spaceRow := keyStyle.Render("[space/shift+←→]") + " " + helpStyle.Render(viewLabel)
	b.WriteString(spaceRow + "\n")

	bindings := [][2]string{
		{"←→↑↓", "Navigate"}, {"b", m.settings.DefaultOffice}, {"f", m.settings.FlexCredit},
		{"n/p", "Next/Prev period"}, {"a", "Add event"}, {"d", "Delete event"},
		{"s", "Search"}, {"w", "What-if"}, {"g", "Git backup"},
		{"v", "Vacations"}, {"h", "Holidays"}, {"o", "Settings"},
		{"q", "Quit"},
	}

	const keyColWidth = 24
	const helpCols = 3
	for i := 0; i < len(bindings); i += helpCols {
		var row strings.Builder
		for j := 0; j < helpCols && i+j < len(bindings); j++ {
			entry := bindings[i+j]
			cell := keyStyle.Render("["+entry[0]+"]") + " " + helpStyle.Render(entry[1])
			visW := lipgloss.Width(cell)
			pad := keyColWidth - visW
			if pad < 1 {
				pad = 1
			}
			row.WriteString(cell + strings.Repeat(" ", pad))
		}
		b.WriteString(row.String() + "\n")
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(dimStyle.Render("Data: "+m.dataDir) + "\n")

	if m.gitInfo.IsRepo {
		dirty := m.hasUnsavedChanges()
		var parts []string
		if dirty {
			parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("unsaved changes"))
		}
		if m.gitInfo.Modified > 0 {
			parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("%d modified", m.gitInfo.Modified)))
		}
		if m.gitInfo.Untracked > 0 {
			parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(fmt.Sprintf("%d untracked", m.gitInfo.Untracked)))
		}
		if len(parts) == 0 {
			parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Render("clean"))
		}

		statusLine := "  Git: " + strings.Join(parts, ", ")
		if m.gitInfo.HasRemote {
			statusLine += dimStyle.Render("  (remote: origin)")
		}
		if dirty || m.gitInfo.Modified > 0 || m.gitInfo.Untracked > 0 {
			statusLine += dimStyle.Render("  [press g to backup]")
		}
		b.WriteString(statusLine + "\n")

		if m.gitInfo.LastCommit != "" {
			b.WriteString(dimStyle.Render("  Last: "+m.gitInfo.LastCommit) + "\n")
		}
	}

	return b.String()
}

func (m *AppModel) renderVacations() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	b.WriteString(titleStyle.Render(" Vacations  (a=add  e=edit  x=delete  q=back)") + "\n")

	header := fmt.Sprintf(" %-4s  %-45s  %-12s  %-12s  %s", "#", "Destination", "Start", "End", "Approved")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(header) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" ----  ---------------------------------------------  ------------  ------------  --------") + "\n")

	all := m.vacationData.All()
	for i, v := range all {
		approved := "No"
		if v.Approved {
			approved = "Yes"
		}
		line := fmt.Sprintf(" %-4d  %-45s  %-12s  %-12s  %s",
			i+1, truncateStr(v.Destination, 45), v.StartDate, v.EndDate, approved)
		style := lipgloss.NewStyle()
		if i == m.listCursor && m.mode == ModeNormal {
			style = style.Reverse(true)
		}
		b.WriteString(style.Render(line) + "\n")
	}

	if m.mode == ModeAdd || m.mode == ModeEdit {
		b.WriteString("\n")
		formTitle := "Add Vacation"
		if m.mode == ModeEdit {
			formTitle = "Edit Vacation"
		}
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(formTitle) + "\n")

		fields := []string{"Destination", "Start Date (YYYY-MM-DD)", "End Date (YYYY-MM-DD)", "Approved (y/n)"}
		for i, field := range fields {
			style := lipgloss.NewStyle()
			if i == m.formCursor {
				style = style.Foreground(lipgloss.Color("226"))
			}
			val := ""
			if m.formInputs != nil && i < len(m.formInputs) {
				val = m.formInputs[i]
			}
			suffix := ""
			if i == m.formCursor {
				suffix = "_"
			}
			b.WriteString(style.Render(fmt.Sprintf("  %s: %s%s\n", field, val, suffix)))
		}
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  (Enter=next field, Esc=cancel)") + "\n")
	}

	return b.String()
}

func (m *AppModel) renderHolidays() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	b.WriteString(titleStyle.Render(" Holidays  (a=add  e=edit  x=delete  q=back)") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" Date          Name") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" ------------  ------------------------------------------") + "\n")

	all := m.holidayData.All()
	for i, h := range all {
		line := fmt.Sprintf(" %-12s  %-42s", h.Date, h.Name)
		style := lipgloss.NewStyle()
		if i == m.listCursor && m.mode == ModeNormal {
			style = style.Reverse(true)
		}
		b.WriteString(style.Render(line) + "\n")
	}

	if m.mode == ModeAdd || m.mode == ModeEdit {
		b.WriteString("\n")
		formTitle := "Add Holiday"
		if m.mode == ModeEdit {
			formTitle = "Edit Holiday"
		}
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(formTitle) + "\n")

		fields := []string{"Date (YYYY-MM-DD)", "Name"}
		for i, field := range fields {
			style := lipgloss.NewStyle()
			if i == m.formCursor {
				style = style.Foreground(lipgloss.Color("226"))
			}
			val := ""
			if m.formInputs != nil && i < len(m.formInputs) {
				val = m.formInputs[i]
			}
			suffix := ""
			if i == m.formCursor {
				suffix = "_"
			}
			b.WriteString(style.Render(fmt.Sprintf("  %s: %s%s\n", field, val, suffix)))
		}
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  (Enter=next field, Esc=cancel)") + "\n")
	}

	return b.String()
}

func (m *AppModel) renderSettings() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	b.WriteString(titleStyle.Render(" Settings  (↑↓=select  Enter/e=edit  q=back)") + "\n\n")

	settings := [][2]string{
		{"Default Office", m.settings.DefaultOffice},
		{"Flex Credit Label", m.settings.FlexCredit},
		{"Goal (%)", fmt.Sprintf("%d", m.settings.Goal)},
	}

	for i, s := range settings {
		style := lipgloss.NewStyle()
		if i == m.listCursor {
			style = style.Reverse(true)
		}
		val := s[1]
		if m.mode == ModeEdit && i == m.listCursor {
			val = m.inputBuffer + "_"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		}
		b.WriteString(style.Render(fmt.Sprintf("  %-20s  %s\n", s[0]+":", val)))
	}

	if m.mode == ModeEdit {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  (Enter=save  Esc=cancel)") + "\n")
	}

	return b.String()
}

func (m *AppModel) renderYearStatsContent() string {
	ys := m.yearStats
	if ys == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(renderStatRow("  Total Calendar Days", fmt.Sprintf("%d", ys.TotalCalendarDays), "") + "\n")
	b.WriteString(renderStatRow("  Total Working Days", fmt.Sprintf("%d", ys.AvailableWorkdays-ys.Holidays), "") + "\n")
	b.WriteString(renderStatRow("  Available Working Days", fmt.Sprintf("%d", ys.TotalDays), "") + "\n")
	b.WriteString(renderStatRow("  Holidays", fmt.Sprintf("%d", ys.Holidays), "") + "\n")
	b.WriteString(renderStatRow("  Vacation Days", fmt.Sprintf("%d", ys.VacationDays), "") + "\n")
	b.WriteString(renderStatRow("  Office Days", fmt.Sprintf("%d", ys.DaysBadgedIn), "") + "\n")
	badgeOnly := ys.DaysBadgedIn - ys.FlexDays
	yBadgePct := ""
	yFlexPct := ""
	if ys.DaysBadgedIn > 0 {
		yBadgePct = fmt.Sprintf("%.1f%%", float64(badgeOnly)/float64(ys.DaysBadgedIn)*100)
		yFlexPct = fmt.Sprintf("%.1f%%", float64(ys.FlexDays)/float64(ys.DaysBadgedIn)*100)
	}
	b.WriteString(renderStatRow("   Badge-In Days", fmt.Sprintf("%d", badgeOnly), yBadgePct) + "\n")
	b.WriteString(renderStatRow("   Flex Credits", fmt.Sprintf("%d", ys.FlexDays), yFlexPct) + "\n")

	return b.String()
}

func (m *AppModel) renderYearStats() string {
	content := m.renderYearStatsContent()
	if content == "" {
		return "No year stats available"
	}
	year := m.today.Year()
	currentPeriod, _ := m.timePeriodData.GetPeriodByDate(m.navDate)
	if currentPeriod != nil {
		year = currentPeriod.StartDate.Year()
	}
	return renderBoxWithTitle(fmt.Sprintf("Year Stats: %d", year), content)
}
