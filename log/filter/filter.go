package filter

import (
	"github.com/noxyicm/wsf/log/event"

	"github.com/pkg/errors"
)

var (
	buildHandlers = map[string]func(map[string]interface{}) (Interface, error){}
)

// Interface represetns message filter for writer
type Interface interface {
	Accept(*event.Event) bool
}

// NewFilter creates a new filter specified by type
func NewFilter(options map[string]interface{}) (Interface, error) {
	var filterType string
	if v, ok := options["type"]; ok {
		if v, ok := v.(string); ok {
			filterType = v
		} else {
			return nil, errors.New("Filter type must be string")
		}
	}

	if f, ok := buildHandlers[filterType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized filter type \"%v\"", filterType)
}

// Register registers a handler for writer filter creation
func Register(filterType string, handler func(map[string]interface{}) (Interface, error)) {
	buildHandlers[filterType] = handler
}
