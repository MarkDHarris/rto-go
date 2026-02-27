package data

import (
	"sort"
)

const eventsFilename = "events.json"

// Event represents a calendar event for a specific date.
type Event struct {
	Date        string `json:"date"`
	Description string `json:"description"`
}

type eventDataFile struct {
	Events []Event `json:"events"`
}

// EventData is the in-memory container for events.
type EventData struct {
	events []Event
}

// NewEventData creates an empty EventData.
func NewEventData() *EventData {
	return &EventData{}
}

// LoadEventData reads event data from the global data directory.
func LoadEventData() (*EventData, error) {
	return LoadEventDataFrom(GetDataDir())
}

// LoadEventDataFrom reads event data from the specified directory.
func LoadEventDataFrom(dir string) (*EventData, error) {
	var file eventDataFile
	if err := LoadJSONFrom(dir, eventsFilename, &file); err != nil {
		return nil, err
	}
	if file.Events == nil {
		file.Events = []Event{}
	}
	return &EventData{events: file.Events}, nil
}

// Save writes event data to the global data directory.
func (e *EventData) Save() error {
	return e.SaveTo(GetDataDir())
}

// SaveTo writes event data to the specified directory.
func (e *EventData) SaveTo(dir string) error {
	file := eventDataFile{Events: e.events}
	return SaveJSONTo(dir, eventsFilename, &file)
}

// Add appends an event and sorts by date.
func (e *EventData) Add(event Event) {
	e.events = append(e.events, event)
	sort.Slice(e.events, func(i, j int) bool {
		return e.events[i].Date < e.events[j].Date
	})
}

// Remove deletes events matching both date and description.
func (e *EventData) Remove(date, description string) {
	filtered := e.events[:0]
	for _, ev := range e.events {
		if ev.Date == date && ev.Description == description {
			continue
		}
		filtered = append(filtered, ev)
	}
	e.events = filtered
}

// All returns a copy of all events.
func (e *EventData) All() []Event {
	result := make([]Event, len(e.events))
	copy(result, e.events)
	return result
}

// Len returns the number of events.
func (e *EventData) Len() int {
	return len(e.events)
}

// GetEventMap returns a map of date key → slice of events for that date.
func (e *EventData) GetEventMap() map[string][]Event {
	m := make(map[string][]Event)
	for _, ev := range e.events {
		m[ev.Date] = append(m[ev.Date], ev)
	}
	return m
}
