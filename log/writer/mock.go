package writer

import (
	"wsf/log/event"
)

const (
	// TYPEMock represents mock writer
	TYPEMock = "mock"
)

func init() {
	Register(TYPEMock, NewMockWriter)
}

// Mock writes log events to slice
type Mock struct {
	writer
	events []*event.Event
}

// Write writes message to log
func (w *Mock) Write(e *event.Event) error {
	for _, filter := range w.filters {
		if !filter.Accept(e) {
			return nil
		}
	}

	w.events = append(w.events, e)
	return nil
}

// Shutdown performs activites such as closing open resources
func (w *Mock) Shutdown() {
}

// NewMockWriter creates mock writer
func NewMockWriter(options *Config) (Interface, error) {
	w := &Mock{}
	return w, nil
}
