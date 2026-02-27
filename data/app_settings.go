package data

const settingsFilename = "settings.yaml"
const settingsDefaultOffice = "McLean, VA"
const settingsDefaultFlex = "Flex Credit"
const settingsDefaultGoal = 50
const defaultTimePeriodFile = "workday-fiscal-quarters.yaml"

// AppSettings holds application-level configuration.
type AppSettings struct {
	DefaultOffice string   `yaml:"default_office"`
	FlexCredit    string   `yaml:"flex_credit"`
	Goal          int      `yaml:"goal"`
	TimePeriods   []string `yaml:"time_periods"`
}

// DefaultAppSettings returns settings with sensible defaults.
func DefaultAppSettings() AppSettings {
	return AppSettings{
		DefaultOffice: settingsDefaultOffice,
		FlexCredit:    settingsDefaultFlex,
		Goal:          settingsDefaultGoal,
		TimePeriods:   []string{defaultTimePeriodFile},
	}
}

// LoadAppSettings reads settings from the global data directory.
func LoadAppSettings() (*AppSettings, error) {
	return LoadAppSettingsFrom(GetDataDir())
}

// LoadAppSettingsFrom reads settings from the specified directory.
func LoadAppSettingsFrom(dir string) (*AppSettings, error) {
	s := DefaultAppSettings()
	var loaded AppSettings
	if err := LoadYAMLFrom(dir, settingsFilename, &loaded); err != nil {
		return nil, err
	}
	if loaded.DefaultOffice != "" {
		s.DefaultOffice = loaded.DefaultOffice
	}
	if loaded.FlexCredit != "" {
		s.FlexCredit = loaded.FlexCredit
	}
	if loaded.Goal > 0 {
		s.Goal = loaded.Goal
	}
	if len(loaded.TimePeriods) > 0 {
		s.TimePeriods = loaded.TimePeriods
	}
	return &s, nil
}

// Save writes settings to the global data directory.
func (s *AppSettings) Save() error {
	return s.SaveTo(GetDataDir())
}

// SaveTo writes settings to the specified directory.
func (s *AppSettings) SaveTo(dir string) error {
	return SaveYAMLTo(dir, settingsFilename, s)
}

// ActiveTimePeriodFile returns the filename for the given index (0-based) in
// the TimePeriods list, falling back to the default if out of range.
func (s *AppSettings) ActiveTimePeriodFile(idx int) string {
	if idx >= 0 && idx < len(s.TimePeriods) {
		return s.TimePeriods[idx]
	}
	if len(s.TimePeriods) > 0 {
		return s.TimePeriods[0]
	}
	return defaultTimePeriodFile
}
