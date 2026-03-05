package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"rto/backup"
	"rto/data"
)

func (m *AppModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.statusMsg = ""

	switch m.currentView {
	case ViewCalendar:
		return m.handleCalendarKey(msg)
	case ViewVacations:
		return m.handleVacationsKey(msg)
	case ViewHolidays:
		return m.handleHolidaysKey(msg)
	case ViewSettings:
		return m.handleSettingsKey(msg)
	}
	return m, nil
}

func (m *AppModel) handleCalendarKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.mode == ModeAdd {
		return m.handleAddEventKey(msg)
	}
	if m.mode == ModeDelete {
		return m.handleDeleteEventKey(msg)
	}
	if m.mode == ModeSearch {
		return m.handleSearchKey(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if m.isWhatIf() {
			m.exitWhatIf()
			return m, nil
		}
		return m, tea.Quit
	case "b":
		m.toggleBadge()
		return m, nil
	case "f":
		m.toggleFlex()
		return m, nil
	case "g":
		m.gitBackup()
		return m, nil
	case "w":
		if m.isWhatIf() {
			m.exitWhatIf()
		} else {
			m.enterWhatIf()
		}
		return m, nil
	case "n":
		m.navigateToAdjacentPeriod(1)
		return m, nil
	case "p":
		m.navigateToAdjacentPeriod(-1)
		return m, nil
	case "a":
		m.mode = ModeAdd
		m.inputBuffer = ""
		m.formCursor = 0
	case "d":
		m.mode = ModeDelete
		m.inputBuffer = ""
		m.formCursor = 0
	case "s":
		m.mode = ModeSearch
		m.inputBuffer = ""
		m.formCursor = 0
	case "v":
		m.currentView = ViewVacations
		m.listCursor = 0
		m.mode = ModeNormal
	case "h":
		m.currentView = ViewHolidays
		m.listCursor = 0
		m.mode = ModeNormal
	case "o":
		m.currentView = ViewSettings
		m.listCursor = 0
		m.mode = ModeNormal
	case "y":
		// Toggle year stats view (rendered in calendar view already)
	case "space":
		m.switchTimePeriodView(1)
		return m, nil
	case "shift+right":
		m.switchTimePeriodView(1)
		return m, nil
	case "shift+left":
		m.switchTimePeriodView(-1)
		return m, nil
	case "right":
		m.selectedDate = m.selectedDate.AddDate(0, 0, 1)
		m.ensureNavFollowsDate()
	case "left":
		m.selectedDate = m.selectedDate.AddDate(0, 0, -1)
		m.ensureNavFollowsDate()
	case "down":
		m.selectedDate = m.selectedDate.AddDate(0, 0, 7)
		m.ensureNavFollowsDate()
	case "up":
		m.selectedDate = m.selectedDate.AddDate(0, 0, -7)
		m.ensureNavFollowsDate()
	}
	return m, nil
}

func (m *AppModel) handleAddEventKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
	case "enter":
		if strings.TrimSpace(m.inputBuffer) != "" {
			m.eventData.Add(data.Event{
				Date:        m.selectedDate.Format("2006-01-02"),
				Description: strings.TrimSpace(m.inputBuffer),
			})
			m.markDirty()
		}
		m.mode = ModeNormal
		m.inputBuffer = ""
	case "backspace":
		if m.formCursor > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:m.formCursor-1]) + string(runes[m.formCursor:])
			m.formCursor--
		}
	case "left":
		if m.formCursor > 0 {
			m.formCursor--
		}
	case "right":
		if m.formCursor < len(m.inputBuffer) {
			m.formCursor++
		}
	default:
		if msg.Text != "" {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:m.formCursor]) + msg.Text + string(runes[m.formCursor:])
			m.formCursor += len([]rune(msg.Text))
		}
	}
	return m, nil
}

func (m *AppModel) handleDeleteEventKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
	case "enter":
		if strings.TrimSpace(m.inputBuffer) != "" {
			m.eventData.Remove(
				m.selectedDate.Format("2006-01-02"),
				strings.TrimSpace(m.inputBuffer),
			)
			m.markDirty()
		}
		m.mode = ModeNormal
		m.inputBuffer = ""
	case "backspace":
		if m.formCursor > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:m.formCursor-1]) + string(runes[m.formCursor:])
			m.formCursor--
		}
	default:
		if msg.Text != "" {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:m.formCursor]) + msg.Text + string(runes[m.formCursor:])
			m.formCursor += len([]rune(msg.Text))
		}
	}
	return m, nil
}

func (m *AppModel) handleSearchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
	case "enter":
		// m.searchResults = searchEvents(m.eventData.All(), m.inputBuffer)
	case "backspace":
		if m.formCursor > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:m.formCursor-1]) + string(runes[m.formCursor:])
			m.formCursor--
		}
	default:
		if msg.Text != "" {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:m.formCursor]) + msg.Text + string(runes[m.formCursor:])
			m.formCursor += len([]rune(msg.Text))
		}
	}
	return m, nil
}

func (m *AppModel) handleVacationsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleVacationsNormal(msg)
	case ModeAdd, ModeEdit:
		return m.handleVacationsForm(msg)
	}
	return m, nil
}

func (m *AppModel) handleVacationsNormal(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	all := m.vacationData.All()
	switch msg.String() {
	case "q":
		m.currentView = ViewCalendar
	case "a":
		m.mode = ModeAdd
		m.formCursor = 0
		m.formInputs = []string{"", "", "", ""}
	case "e", "enter":
		if len(all) > 0 {
			v := all[m.listCursor]
			approved := "n"
			if v.Approved {
				approved = "y"
			}
			m.mode = ModeEdit
			m.formCursor = 0
			m.formInputs = []string{v.Destination, v.StartDate, v.EndDate, approved}
		}
	case "x":
		if len(all) > 0 {
			v := all[m.listCursor]
			m.vacationData.Remove(v.StartDate, v.EndDate)
			m.markDirty()
			if m.listCursor >= m.vacationData.Len() && m.listCursor > 0 {
				m.listCursor--
			}
		}
	case "down":
		if m.listCursor < len(all)-1 {
			m.listCursor++
		}
	case "up":
		if m.listCursor > 0 {
			m.listCursor--
		}
	}

	return m, nil
}

func (m *AppModel) handleVacationsForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.formInputs = nil
	case "enter":
		if m.formCursor < len(m.formInputs)-1 {
			m.formCursor++
		} else {
			approved := strings.ToLower(strings.TrimSpace(m.formInputs[3])) == "y"
			newVac := data.Vacation{
				Destination: strings.TrimSpace(m.formInputs[0]),
				StartDate:   strings.TrimSpace(m.formInputs[1]),
				EndDate:     strings.TrimSpace(m.formInputs[2]),
				Approved:    approved,
			}
			if m.mode == ModeEdit {
				all := m.vacationData.All()
				if m.listCursor < len(all) {
					old := all[m.listCursor]
					m.vacationData.Remove(old.StartDate, old.EndDate)
				}
			}
			m.vacationData.Add(newVac)
			m.markDirty()
			m.mode = ModeNormal
			m.formInputs = nil
			m.recalculateStats()
		}
	case "backspace":
		if len(m.formInputs[m.formCursor]) > 0 {
			runes := []rune(m.formInputs[m.formCursor])
			m.formInputs[m.formCursor] = string(runes[:len(runes)-1])
		}
	case "tab":
		if m.formCursor < len(m.formInputs)-1 {
			m.formCursor++
		}
	default:
		if msg.Text != "" {
			m.formInputs[m.formCursor] += msg.Text
		}
	}
	return m, nil
}

func (m *AppModel) handleHolidaysKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleHolidaysNormal(msg)
	case ModeAdd, ModeEdit:
		return m.handleHolidaysForm(msg)
	}
	return m, nil
}

func (m *AppModel) handleHolidaysNormal(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	all := m.holidayData.All()
	switch msg.String() {
	case "q":
		m.currentView = ViewCalendar
	case "a":
		m.mode = ModeAdd
		m.formCursor = 0
		m.formInputs = []string{"", ""}
	case "e", "enter":
		if len(all) > 0 {
			h := all[m.listCursor]
			m.mode = ModeEdit
			m.formCursor = 0
			m.formInputs = []string{h.Date, h.Name}
		}
	case "x":
		if len(all) > 0 {
			h := all[m.listCursor]
			newHD := data.NewHolidayData()
			for _, existing := range all {
				if existing.Date == h.Date && existing.Name == h.Name {
					continue
				}
				newHD.Add(existing)
			}
			*m.holidayData = *newHD
			m.markDirty()
			if m.listCursor >= m.holidayData.Len() && m.listCursor > 0 {
				m.listCursor--
			}
			m.recalculateStats()
		}
	case "down":
		if m.listCursor < len(all)-1 {
			m.listCursor++
		}
	case "up":
		if m.listCursor > 0 {
			m.listCursor--
		}
	}
	return m, nil
}

func (m *AppModel) handleHolidaysForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.formInputs = nil
	case "enter":
		if m.formCursor < len(m.formInputs)-1 {
			m.formCursor++
		} else {
			newH := data.Holiday{
				Date: strings.TrimSpace(m.formInputs[0]),
				Name: strings.TrimSpace(m.formInputs[1]),
			}
			if m.mode == ModeEdit {
				all := m.holidayData.All()
				if m.listCursor < len(all) {
					old := all[m.listCursor]
					newHD := data.NewHolidayData()
					for _, h := range all {
						if h.Date == old.Date && h.Name == old.Name {
							continue
						}
						newHD.Add(h)
					}
					*m.holidayData = *newHD
				}
			}
			m.holidayData.Add(newH)
			m.markDirty()
			m.mode = ModeNormal
			m.formInputs = nil
			m.recalculateStats()
		}
	case "backspace":
		if len(m.formInputs[m.formCursor]) > 0 {
			runes := []rune(m.formInputs[m.formCursor])
			m.formInputs[m.formCursor] = string(runes[:len(runes)-1])
		}
	case "tab":
		if m.formCursor < len(m.formInputs)-1 {
			m.formCursor++
		}
	default:
		if msg.Text != "" {
			m.formInputs[m.formCursor] += msg.Text
		}
	}
	return m, nil
}

func (m *AppModel) handleSettingsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleSettingsNormal(msg)
	case ModeEdit:
		return m.handleSettingsEdit(msg)
	}
	return m, nil
}

func (m *AppModel) handleSettingsNormal(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		m.currentView = ViewCalendar
	case "e", "enter":
		m.mode = ModeEdit
		switch m.listCursor {
		case 0:
			m.inputBuffer = m.settings.DefaultOffice
		case 1:
			m.inputBuffer = m.settings.FlexCredit
		case 2:
			m.inputBuffer = fmt.Sprintf("%d", m.settings.Goal)
		}
		m.formCursor = len(m.inputBuffer)
	case "down":
		if m.listCursor < 2 {
			m.listCursor++
		}
	case "up":
		if m.listCursor > 0 {
			m.listCursor--
		}
	}
	return m, nil
}

func (m *AppModel) handleSettingsEdit(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
	case "enter":
		switch m.listCursor {
		case 0:
			m.settings.DefaultOffice = strings.TrimSpace(m.inputBuffer)
		case 1:
			m.settings.FlexCredit = strings.TrimSpace(m.inputBuffer)
		case 2:
			if v, err := strconv.Atoi(strings.TrimSpace(m.inputBuffer)); err == nil && v > 0 && v <= 100 {
				m.settings.Goal = v
				m.recalculateStats()
			}
		}
		m.markDirty()
		m.mode = ModeNormal
	case "backspace":
		if m.formCursor > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:m.formCursor-1]) + string(runes[m.formCursor:])
			m.formCursor--
		}
	case "left":
		if m.formCursor > 0 {
			m.formCursor--
		}
	case "right":
		if m.formCursor < len(m.inputBuffer) {
			m.formCursor++
		}
	default:
		if msg.Text != "" {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:m.formCursor]) + msg.Text + string(runes[m.formCursor:])
			m.formCursor += len([]rune(msg.Text))
		}
	}
	return m, nil
}

//
// Helper functions moved from calendar_view.go
//

func (m *AppModel) toggleBadge() {
	key := m.selectedDate.Format("2006-01-02")
	if m.badgeData.Has(key) {
		existing, _ := m.badgeData.Get(key)
		if !existing.IsFlexCredit {
			m.badgeData.Remove(key)
		}
	} else {
		m.badgeData.Add(data.BadgeEntry{
			EntryDate:  key,
			DateTime:   data.FlexTime{Time: m.selectedDate},
			Office:     m.settings.DefaultOffice,
			IsBadgedIn: true,
		})
	}
	m.markDirty()
	m.recalculateStats()
}

func (m *AppModel) toggleFlex() {
	key := m.selectedDate.Format("2006-01-02")
	if m.badgeData.Has(key) {
		existing, _ := m.badgeData.Get(key)
		if existing.IsFlexCredit {
			m.badgeData.Remove(key)
		}
	} else {
		m.badgeData.Add(data.BadgeEntry{
			EntryDate:    key,
			DateTime:     data.FlexTime{Time: m.selectedDate},
			Office:       m.settings.FlexCredit,
			IsBadgedIn:   true,
			IsFlexCredit: true,
		})
	}
	m.markDirty()
	m.recalculateStats()
}

func (m *AppModel) navigateToAdjacentPeriod(dir int) {
	currentPeriod, err := m.timePeriodData.GetPeriodByDate(m.navDate)
	if err != nil {
		return
	}
	all := m.timePeriodData.All()
	if currentPeriod == nil {
		return
	}
	for i, q := range all {
		if q.Key == currentPeriod.Key {
			next := i + dir
			if next >= 0 && next < len(all) {
				np := all[next]
				m.selectedDate = np.StartDate
				m.navDate = time.Date(np.StartDate.Year(), np.StartDate.Month(), 1, 0, 0, 0, 0, time.UTC)
				m.recalculateStats()
			}
			return
		}
	}
}

func (m *AppModel) ensureNavFollowsDate() {
	currentPeriod, _ := m.timePeriodData.GetPeriodByDate(m.navDate)
	if currentPeriod != nil &&
		!m.selectedDate.Before(currentPeriod.StartDate) &&
		!m.selectedDate.After(currentPeriod.EndDate) {
		return
	}

	newPeriod, err := m.timePeriodData.GetPeriodByDate(m.selectedDate)
	if err != nil {
		newPeriod, err = m.timePeriodData.NearestPeriod(m.selectedDate)
		if err != nil {
			return
		}
	}

	m.navDate = time.Date(newPeriod.StartDate.Year(), newPeriod.StartDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	m.recalculateStats()
}

func searchEvents(events []data.Event, query string) []data.Event {
	if query == "" {
		return nil
	}
	lower := strings.ToLower(query)
	var results []data.Event
	for _, ev := range events {
		if strings.Contains(strings.ToLower(ev.Description), lower) ||
			strings.Contains(ev.Date, query) {
			results = append(results, ev)
		}
	}
	return results
}

func (m *AppModel) gitBackup() {
	result := backup.Perform(m.dataDir, "")
	m.statusMsg = result.Message
	m.refreshGitInfo()
	if !result.IsError {
		m.cleanChecksum = m.dataChecksum()
	}
}
