package data

import (
	"time"
)

const vacationsFilename = "vacations.yaml"

// Vacation represents a single vacation period.
type Vacation struct {
	Destination string `yaml:"destination" json:"destination"`
	StartDate   string `yaml:"start_date" json:"start_date"`
	EndDate     string `yaml:"end_date" json:"end_date"`
	Approved    bool   `yaml:"approved" json:"approved"`
}

type vacationDataFile struct {
	Vacations []Vacation `yaml:"vacations"`
}

// VacationData is the in-memory container for vacations.
type VacationData struct {
	vacations []Vacation
}

// NewVacationData creates an empty VacationData.
func NewVacationData() *VacationData {
	return &VacationData{}
}

// LoadVacationData reads vacation data from the global data directory.
func LoadVacationData() (*VacationData, error) {
	return LoadVacationDataFrom(GetDataDir())
}

// LoadVacationDataFrom reads vacation data from the specified directory.
func LoadVacationDataFrom(dir string) (*VacationData, error) {
	var file vacationDataFile
	if err := LoadYAMLFrom(dir, vacationsFilename, &file); err != nil {
		return nil, err
	}
	if file.Vacations == nil {
		file.Vacations = []Vacation{}
	}
	return &VacationData{vacations: file.Vacations}, nil
}

// Save writes vacation data to the global data directory.
func (v *VacationData) Save() error {
	return v.SaveTo(GetDataDir())
}

// SaveTo writes vacation data to the specified directory.
func (v *VacationData) SaveTo(dir string) error {
	file := vacationDataFile{Vacations: v.vacations}
	return SaveYAMLTo(dir, vacationsFilename, &file)
}

// Add appends a vacation.
func (v *VacationData) Add(vacation Vacation) {
	v.vacations = append(v.vacations, vacation)
}

// Remove deletes vacations matching both start and end date.
func (v *VacationData) Remove(startDate, endDate string) {
	filtered := v.vacations[:0]
	for _, vac := range v.vacations {
		if vac.StartDate == startDate && vac.EndDate == endDate {
			continue
		}
		filtered = append(filtered, vac)
	}
	v.vacations = filtered
}

// All returns a copy of all vacations.
func (v *VacationData) All() []Vacation {
	result := make([]Vacation, len(v.vacations))
	copy(result, v.vacations)
	return result
}

// Len returns the number of vacations.
func (v *VacationData) Len() int {
	return len(v.vacations)
}

// GetVacationMap expands all vacation date ranges into individual date keys.
// Only weekdays (Mon–Fri) are included; holidays are NOT excluded here.
func (v *VacationData) GetVacationMap() map[string]Vacation {
	m := make(map[string]Vacation)
	for _, vac := range v.vacations {
		start, err := time.Parse(BadgeDateFormat, vac.StartDate)
		if err != nil {
			continue
		}
		end, err := time.Parse(BadgeDateFormat, vac.EndDate)
		if err != nil {
			continue
		}
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
				continue
			}
			m[d.Format(BadgeDateFormat)] = vac
		}
	}
	return m
}
