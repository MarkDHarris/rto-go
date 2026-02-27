package app

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	"rto/backup"
	"rto/calc"
	"rto/data"

	tea "github.com/charmbracelet/bubbletea"
)

type ViewState int

const (
	ViewCalendar ViewState = iota
	ViewVacations
	ViewHolidays
	ViewSettings
	ViewYearStats
)

type ViewMode int

const (
	ModeNormal ViewMode = iota
	ModeAdd
	ModeEdit
	ModeDelete
	ModeSearch
)

type AppModel struct {
	// Business data
	timePeriodData *data.TimePeriodData
	badgeData      *data.BadgeEntryData
	holidayData    *data.HolidayData
	vacationData   *data.VacationData
	eventData      *data.EventData
	settings       *data.AppSettings
	dataDir        string

	// UI state
	currentView  ViewState
	mode         ViewMode
	selectedDate time.Time
	navDate      time.Time
	today        time.Time
	listCursor   int
	inputBuffer  string
	formInputs   []string
	formCursor   int

	// Time period view switching
	activeTimePeriodIdx int

	// What-if mode
	whatIfSnapshot      *data.BadgeEntryData
	whatIfDirtySnapshot string

	// Bubbletea helpers
	err           error
	termWidth     int
	termHeight    int
	activeStats   *calc.PeriodStats
	yearStats     *calc.PeriodStats
	statusMsg     string
	gitInfo       backup.StatusInfo
	cleanChecksum string
}

func New() (*AppModel, error) {
	settings, err := data.LoadAppSettings()
	if err != nil {
		return nil, fmt.Errorf("loading settings: %w", err)
	}

	dir := data.GetDataDir()
	tpFile := settings.ActiveTimePeriodFile(0)
	timePeriodData, err := data.LoadTimePeriodDataFrom(dir, tpFile)
	if err != nil {
		return nil, fmt.Errorf("loading time periods (%s): %w", tpFile, err)
	}

	badgeData, err := data.LoadBadgeEntryData()
	if err != nil {
		return nil, fmt.Errorf("loading badge data: %w", err)
	}

	holidayData, err := data.LoadHolidayData()
	if err != nil {
		return nil, fmt.Errorf("loading holidays: %w", err)
	}

	vacationData, err := data.LoadVacationData()
	if err != nil {
		return nil, fmt.Errorf("loading vacations: %w", err)
	}

	eventData, err := data.LoadEventData()
	if err != nil {
		return nil, fmt.Errorf("loading events: %w", err)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	navDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	m := &AppModel{
		timePeriodData:      timePeriodData,
		badgeData:           badgeData,
		holidayData:         holidayData,
		vacationData:        vacationData,
		eventData:           eventData,
		settings:            settings,
		dataDir:             dir,
		currentView:         ViewCalendar,
		mode:                ModeNormal,
		selectedDate:        today,
		navDate:             navDate,
		today:               today,
		activeTimePeriodIdx: 0,
	}
	m.recalculateStats()
	m.refreshGitInfo()
	m.cleanChecksum = m.dataChecksum()
	return m, nil
}

func (m *AppModel) refreshGitInfo() {
	m.gitInfo = backup.Status(m.dataDir)
}

func (m *AppModel) dataChecksum() string {
	h := sha256.New()

	badges := m.badgeData.All()
	sort.Slice(badges, func(i, j int) bool { return badges[i].EntryDate < badges[j].EntryDate })
	for _, b := range badges {
		fmt.Fprintf(h, "B|%s|%v|", b.EntryDate, b.IsFlexCredit)
	}

	events := m.eventData.All()
	sort.Slice(events, func(i, j int) bool {
		if events[i].Date != events[j].Date {
			return events[i].Date < events[j].Date
		}
		return events[i].Description < events[j].Description
	})
	for _, e := range events {
		fmt.Fprintf(h, "E|%s|%s|", e.Date, e.Description)
	}

	vacations := m.vacationData.All()
	sort.Slice(vacations, func(i, j int) bool { return vacations[i].StartDate < vacations[j].StartDate })
	for _, v := range vacations {
		fmt.Fprintf(h, "V|%s|%s|%s|%v|", v.StartDate, v.EndDate, v.Destination, v.Approved)
	}

	holidays := m.holidayData.All()
	sort.Slice(holidays, func(i, j int) bool { return holidays[i].Date < holidays[j].Date })
	for _, hd := range holidays {
		fmt.Fprintf(h, "H|%s|%s|", hd.Date, hd.Name)
	}

	fmt.Fprintf(h, "S|%s|%s|%d|", m.settings.DefaultOffice, m.settings.FlexCredit, m.settings.Goal)

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m *AppModel) hasUnsavedChanges() bool {
	return m.dataChecksum() != m.cleanChecksum
}

func (m *AppModel) markDirty() {
	// no-op; kept for call-site compatibility — hasUnsavedChanges() is checksum-based
}

func (m *AppModel) Init() tea.Cmd {
	return nil
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *AppModel) GetSettings() *data.AppSettings {
	return m.settings
}

func (m *AppModel) BadgeData() *data.BadgeEntryData {
	return m.badgeData
}

func (m *AppModel) EventData() *data.EventData {
	return m.eventData
}

func (m *AppModel) VacationData() *data.VacationData {
	return m.vacationData
}

func (m *AppModel) HolidayData() *data.HolidayData {
	return m.holidayData
}

func (m *AppModel) TimePeriodData() *data.TimePeriodData {
	return m.timePeriodData
}

func (m *AppModel) DataDir() string {
	return m.dataDir
}

func (m *AppModel) recalculateStats() {
	period, err := m.timePeriodData.GetPeriodByDate(m.navDate)
	if err != nil {
		return
	}
	stats, err := calc.CalculatePeriodStats(
		period,
		m.badgeData,
		m.holidayData,
		m.vacationData,
		m.settings.Goal,
		&m.today,
	)
	if err != nil {
		return
	}
	m.activeStats = stats
	m.recalculateYearStats(period)
}

func (m *AppModel) recalculateYearStats(period *data.TimePeriod) {
	year := m.today.Year()
	if period != nil {
		year = period.StartDate.Year()
	}

	all := m.timePeriodData.All()
	var yearPeriods []*data.TimePeriod
	for i := range all {
		if all[i].StartDate.Year() == year {
			p := all[i]
			yearPeriods = append(yearPeriods, &p)
		}
	}
	if len(yearPeriods) == 0 {
		return
	}

	stats, err := calc.CalculateYearStats(yearPeriods, m.badgeData, m.holidayData, m.vacationData, m.settings.Goal, &m.today)
	if err == nil {
		m.yearStats = stats
	}
}

func (m *AppModel) switchTimePeriodView(dir int) {
	n := len(m.settings.TimePeriods)
	if n <= 1 {
		return
	}
	m.activeTimePeriodIdx = (m.activeTimePeriodIdx + dir + n) % n
	tpFile := m.settings.ActiveTimePeriodFile(m.activeTimePeriodIdx)
	td, err := data.LoadTimePeriodDataFrom(m.dataDir, tpFile)
	if err != nil {
		m.statusMsg = fmt.Sprintf("Error loading %s: %v", tpFile, err)
		return
	}
	m.timePeriodData = td
	m.statusMsg = fmt.Sprintf("View: %s", tpFile)

	now := time.Now()
	m.navDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	m.selectedDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if p, err := m.timePeriodData.GetPeriodByDate(m.selectedDate); err == nil {
		m.navDate = time.Date(p.StartDate.Year(), p.StartDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else if p, err := m.timePeriodData.NearestPeriod(m.selectedDate); err == nil {
		m.navDate = time.Date(p.StartDate.Year(), p.StartDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		m.selectedDate = p.StartDate
	}
	m.recalculateStats()
}

func (m *AppModel) isWhatIf() bool {
	return m.whatIfSnapshot != nil
}

func (m *AppModel) enterWhatIf() {
	m.whatIfSnapshot = m.badgeData.Clone()
	m.whatIfDirtySnapshot = m.cleanChecksum
}

func (m *AppModel) exitWhatIf() {
	if m.whatIfSnapshot != nil {
		m.badgeData = m.whatIfSnapshot
		m.cleanChecksum = m.whatIfDirtySnapshot
		m.whatIfSnapshot = nil
		m.recalculateStats()
	}
}
