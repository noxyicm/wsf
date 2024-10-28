package writer

import (
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log/event"
	"github.com/noxyicm/wsf/log/filter"
	"github.com/noxyicm/wsf/log/formatter"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

// Interface represents message writer
type Interface interface {
	AddFilter(filter.Interface) error
	SetFormatter(formatter.Interface) error
	GetFormatter() formatter.Interface
	Write(*event.Event) error
	Shutdown()
}

type writer struct {
	Enable    bool
	Filters   []filter.Interface
	Formatter formatter.Interface
}

// AddFilter adds filter to writer
func (w *writer) AddFilter(f filter.Interface) error {
	w.Filters = append(w.Filters, f)
	return nil
}

// SetFormatter sets formater for this writer
func (w *writer) SetFormatter(frmt formatter.Interface) error {
	w.Formatter = frmt
	return nil
}

// Formatter returns writer formatter
func (w *writer) GetFormatter() formatter.Interface {
	return w.Formatter
}

// NewWriter creates a new writer specified by type
func NewWriter(options *Config) (Interface, error) {
	var writerType string
	if v, ok := options.Params["type"]; ok {
		if v, ok := v.(string); ok {
			writerType = v
		} else {
			return nil, errors.New("Writer type must be string")
		}
	}

	if f, ok := buildHandlers[writerType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized writer type '%v'", writerType)
}

// Register registers a handler for writer creation
func Register(writerType string, handler func(*Config) (Interface, error)) {
	buildHandlers[writerType] = handler
}
