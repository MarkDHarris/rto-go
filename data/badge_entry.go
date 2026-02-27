package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const badgeDataFilename = "badge_data.json"

// BadgeDateFormat is the canonical date key format (matches YYYY-MM-DD).
const BadgeDateFormat = "2006-01-02"

// flexTimeFormats lists the datetime formats we accept on load, in order of preference.
var flexTimeFormats = []string{
	time.RFC3339,           // "2006-01-02T15:04:05Z07:00"
	"2006-01-02T15:04:05",  // naive (no timezone) — used by original rto data files
	"2006-01-02T15:04:05Z", // explicit UTC Z
	BadgeDateFormat,        // date-only fallback
}

// FlexTime wraps time.Time and accepts both RFC3339 and timezone-less datetime strings.
type FlexTime struct{ time.Time }

func (ft *FlexTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	for _, f := range flexTimeFormats {
		if t, err := time.Parse(f, s); err == nil {
			ft.Time = t
			return nil
		}
	}
	return fmt.Errorf("cannot parse datetime %q", s)
}

func (ft FlexTime) MarshalJSON() ([]byte, error) {
	// Always write in the naive format to stay compatible with rto data files.
	return json.Marshal(ft.Time.Format("2006-01-02T15:04:05"))
}

// BadgeEntry represents a single badge-in event.
type BadgeEntry struct {
	EntryDate    string   `json:"entry_date"`
	DateTime     FlexTime `json:"date_time"`
	Office       string   `json:"office"`
	IsBadgedIn   bool     `json:"is_badged_in"`
	IsFlexCredit bool     `json:"is_flex_credit"`
}

type badgeDataFile struct {
	BadgeData []BadgeEntry `json:"badge_data"`
}

// BadgeEntryData is the in-memory container for all badge entries.
type BadgeEntryData struct {
	entries []BadgeEntry
}

// NewBadgeEntryData creates an empty BadgeEntryData.
func NewBadgeEntryData() *BadgeEntryData {
	return &BadgeEntryData{}
}

// Load reads badge data from the global data directory.
func LoadBadgeEntryData() (*BadgeEntryData, error) {
	return LoadBadgeEntryDataFrom(GetDataDir())
}

// LoadBadgeEntryDataFrom reads badge data from the specified directory.
func LoadBadgeEntryDataFrom(dir string) (*BadgeEntryData, error) {
	var file badgeDataFile
	if err := LoadJSONFrom(dir, badgeDataFilename, &file); err != nil {
		return nil, err
	}
	if file.BadgeData == nil {
		file.BadgeData = []BadgeEntry{}
	}
	return &BadgeEntryData{entries: file.BadgeData}, nil
}

// Save writes badge data to the global data directory.
func (b *BadgeEntryData) Save() error {
	return b.SaveTo(GetDataDir())
}

// SaveTo writes badge data to the specified directory.
func (b *BadgeEntryData) SaveTo(dir string) error {
	file := badgeDataFile{BadgeData: b.entries}
	return SaveJSONTo(dir, badgeDataFilename, &file)
}

// Has returns true if a badge entry exists for the given date key (YYYY-MM-DD).
func (b *BadgeEntryData) Has(key string) bool {
	for _, e := range b.entries {
		if e.EntryDate == key {
			return true
		}
	}
	return false
}

// Get returns the badge entry for the given date key, if it exists.
func (b *BadgeEntryData) Get(key string) (BadgeEntry, bool) {
	for _, e := range b.entries {
		if e.EntryDate == key {
			return e, true
		}
	}
	return BadgeEntry{}, false
}

// Add appends a new badge entry.
func (b *BadgeEntryData) Add(entry BadgeEntry) {
	b.entries = append(b.entries, entry)
}

// Remove deletes the badge entry for the given date key.
func (b *BadgeEntryData) Remove(key string) {
	filtered := b.entries[:0]
	for _, e := range b.entries {
		if e.EntryDate != key {
			filtered = append(filtered, e)
		}
	}
	b.entries = filtered
}

// Len returns the number of entries.
func (b *BadgeEntryData) Len() int {
	return len(b.entries)
}

// All returns a copy of all entries.
func (b *BadgeEntryData) All() []BadgeEntry {
	result := make([]BadgeEntry, len(b.entries))
	copy(result, b.entries)
	return result
}

// GetBadgeMap returns a map of date key → BadgeEntry for entries in [start, end].
func (b *BadgeEntryData) GetBadgeMap(start, end time.Time) map[string]BadgeEntry {
	m := make(map[string]BadgeEntry)
	for _, entry := range b.entries {
		t, err := time.Parse(BadgeDateFormat, entry.EntryDate)
		if err != nil {
			continue
		}
		if !t.Before(start) && !t.After(end) {
			m[entry.EntryDate] = entry
		}
	}
	return m
}

// Clone returns a deep copy of the BadgeEntryData.
func (b *BadgeEntryData) Clone() *BadgeEntryData {
	entries := make([]BadgeEntry, len(b.entries))
	copy(entries, b.entries)
	return &BadgeEntryData{entries: entries}
}
