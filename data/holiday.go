package data

const holidaysFilename = "holidays.yaml"

// Holiday represents a single holiday.
type Holiday struct {
	Name string `yaml:"name" json:"name"`
	Date string `yaml:"date" json:"date"`
}

type holidayDataFile struct {
	Holidays []Holiday `yaml:"holidays"`
}

// HolidayData is the in-memory container for holidays.
type HolidayData struct {
	holidays []Holiday
}

// NewHolidayData creates an empty HolidayData.
func NewHolidayData() *HolidayData {
	return &HolidayData{}
}

// LoadHolidayData reads holiday data from the global data directory.
func LoadHolidayData() (*HolidayData, error) {
	return LoadHolidayDataFrom(GetDataDir())
}

// LoadHolidayDataFrom reads holiday data from the specified directory.
func LoadHolidayDataFrom(dir string) (*HolidayData, error) {
	var file holidayDataFile
	if err := LoadYAMLFrom(dir, holidaysFilename, &file); err != nil {
		return nil, err
	}
	if file.Holidays == nil {
		file.Holidays = []Holiday{}
	}
	return &HolidayData{holidays: file.Holidays}, nil
}

// Save writes holiday data to the global data directory.
func (h *HolidayData) Save() error {
	return h.SaveTo(GetDataDir())
}

// SaveTo writes holiday data to the specified directory.
func (h *HolidayData) SaveTo(dir string) error {
	file := holidayDataFile{Holidays: h.holidays}
	return SaveYAMLTo(dir, holidaysFilename, &file)
}

// Add appends a holiday.
func (h *HolidayData) Add(holiday Holiday) {
	h.holidays = append(h.holidays, holiday)
}

// All returns a copy of all holidays.
func (h *HolidayData) All() []Holiday {
	result := make([]Holiday, len(h.holidays))
	copy(result, h.holidays)
	return result
}

// Len returns the number of holidays.
func (h *HolidayData) Len() int {
	return len(h.holidays)
}

// GetHolidayMap returns a map of date key → Holiday.
func (h *HolidayData) GetHolidayMap() map[string]Holiday {
	m := make(map[string]Holiday)
	for _, holiday := range h.holidays {
		m[holiday.Date] = holiday
	}
	return m
}
