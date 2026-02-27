package data

import (
	"fmt"
	"strings"
	"time"
)

const defaultTimePeriodsFilename = "workday-fiscal-quarters.yaml"
const defaultCalendarDisplayColumns = 3

// TimePeriod holds the configuration for a single time period (arbitrary date range).
type TimePeriod struct {
	Key          string `yaml:"key"`
	Name         string `yaml:"name"`
	StartDateRaw string `yaml:"start_date"`
	EndDateRaw   string `yaml:"end_date"`

	StartDate time.Time `yaml:"-"`
	EndDate   time.Time `yaml:"-"`
}

func (tp *TimePeriod) ParseDates() error {
	var err error
	tp.StartDate, err = time.Parse(BadgeDateFormat, tp.StartDateRaw)
	if err != nil {
		return fmt.Errorf("parsing start_date %q for %s: %w", tp.StartDateRaw, tp.Key, err)
	}
	tp.EndDate, err = time.Parse(BadgeDateFormat, tp.EndDateRaw)
	if err != nil {
		return fmt.Errorf("parsing end_date %q for %s: %w", tp.EndDateRaw, tp.Key, err)
	}
	return nil
}

type timePeriodDataFile struct {
	CalendarDisplayColumns int          `yaml:"calendar_display_columns,omitempty"`
	TimePeriods            []TimePeriod `yaml:"timeperiods"`
}

// TimePeriodData is the in-memory container for time period configurations.
type TimePeriodData struct {
	periods                []TimePeriod
	filename               string
	calendarDisplayColumns int
}

func NewTimePeriodData() *TimePeriodData {
	return &TimePeriodData{
		filename:               defaultTimePeriodsFilename,
		calendarDisplayColumns: defaultCalendarDisplayColumns,
	}
}

// NewTimePeriodDataWithFile creates a container that will read/write the
// specified YAML filename rather than the default.
func NewTimePeriodDataWithFile(filename string) *TimePeriodData {
	if filename == "" {
		filename = defaultTimePeriodsFilename
	}
	return &TimePeriodData{
		filename:               filename,
		calendarDisplayColumns: defaultCalendarDisplayColumns,
	}
}

// Filename returns the underlying YAML filename for this data set.
func (td *TimePeriodData) Filename() string {
	return td.filename
}

// CalendarDisplayColumns returns the number of calendar columns for this file.
func (td *TimePeriodData) CalendarDisplayColumns() int {
	if td.calendarDisplayColumns <= 0 {
		return defaultCalendarDisplayColumns
	}
	return td.calendarDisplayColumns
}

// SetCalendarDisplayColumns sets the number of calendar columns.
func (td *TimePeriodData) SetCalendarDisplayColumns(cols int) {
	td.calendarDisplayColumns = cols
}

func LoadTimePeriodData() (*TimePeriodData, error) {
	dir := GetDataDir()
	settings, err := LoadAppSettingsFrom(dir)
	if err != nil {
		return LoadTimePeriodDataFrom(dir, "")
	}
	return LoadTimePeriodDataFrom(dir, settings.ActiveTimePeriodFile(0))
}

func LoadTimePeriodDataFrom(dir string, filename string) (*TimePeriodData, error) {
	if filename == "" {
		filename = defaultTimePeriodsFilename
	}
	var file timePeriodDataFile
	if err := LoadYAMLFrom(dir, filename, &file); err != nil {
		return nil, err
	}
	for i := range file.TimePeriods {
		if err := file.TimePeriods[i].ParseDates(); err != nil {
			return nil, err
		}
	}
	cols := file.CalendarDisplayColumns
	if cols <= 0 {
		cols = defaultCalendarDisplayColumns
	}
	return &TimePeriodData{
		periods:                file.TimePeriods,
		filename:               filename,
		calendarDisplayColumns: cols,
	}, nil
}

func (td *TimePeriodData) Save() error {
	return td.SaveTo(GetDataDir())
}

func (td *TimePeriodData) SaveTo(dir string) error {
	file := timePeriodDataFile{
		CalendarDisplayColumns: td.calendarDisplayColumns,
		TimePeriods:            td.periods,
	}
	return SaveYAMLTo(dir, td.filename, &file)
}

func (td *TimePeriodData) All() []TimePeriod {
	result := make([]TimePeriod, len(td.periods))
	copy(result, td.periods)
	return result
}

func (td *TimePeriodData) Len() int {
	return len(td.periods)
}

func (td *TimePeriodData) Add(tp TimePeriod) {
	td.periods = append(td.periods, tp)
}

func (td *TimePeriodData) GetCurrentPeriod() (*TimePeriod, error) {
	return td.GetPeriodByDate(time.Now())
}

func (td *TimePeriodData) GetPeriodByDate(date time.Time) (*TimePeriod, error) {
	for i := range td.periods {
		tp := &td.periods[i]
		if !date.Before(tp.StartDate) && !date.After(tp.EndDate) {
			return tp, nil
		}
	}
	return nil, fmt.Errorf("no time period found for date %s", date.Format(BadgeDateFormat))
}

func (td *TimePeriodData) GetPeriodByKey(key string) (*TimePeriod, error) {
	for i := range td.periods {
		if td.periods[i].Key == key {
			return &td.periods[i], nil
		}
	}
	return nil, fmt.Errorf("time period %q not found", key)
}

// NearestPeriod returns the time period closest to the given date.
func (td *TimePeriodData) NearestPeriod(date time.Time) (*TimePeriod, error) {
	if len(td.periods) == 0 {
		return nil, fmt.Errorf("no time periods configured")
	}
	if tp, err := td.GetPeriodByDate(date); err == nil {
		return tp, nil
	}
	closest := &td.periods[0]
	minDiff := absDuration(date.Sub(closest.StartDate))
	for i := range td.periods[1:] {
		tp := &td.periods[i+1]
		d := absDuration(date.Sub(tp.StartDate))
		if d < minDiff {
			minDiff = d
			closest = tp
		}
	}
	return closest, nil
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// MonthSpan returns the number of calendar months covered by the period.
func (tp *TimePeriod) MonthSpan() int {
	startY, startM := tp.StartDate.Year(), int(tp.StartDate.Month())
	endY, endM := tp.EndDate.Year(), int(tp.EndDate.Month())
	return (endY-startY)*12 + (endM - startM) + 1
}

// TimePeriodKey builds a key like "Q1_2025" from a period label and year.
func TimePeriodKey(name string, year int) string {
	return fmt.Sprintf("%s_%d", strings.ToUpper(name), year)
}
