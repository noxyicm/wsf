package formatter

import (
	"regexp"
	"wsf/errors"
	"wsf/log/event"
)

var (
	buildHandlers = map[string]func(map[string]interface{}) (Interface, error){}
	rest          = regexp.MustCompile(`\#(.+)\#`)
)

// Interface represents message formatter
type Interface interface {
	Format(*event.Event) (string, error)
}

// NewFormatter creates a new formatter specified by type
func NewFormatter(options map[string]interface{}) (Interface, error) {
	var formatterType string
	if v, ok := options["type"]; ok {
		if v, ok := v.(string); ok {
			formatterType = v
		} else {
			return nil, errors.New("Formatter type must be string")
		}
	}

	if f, ok := buildHandlers[formatterType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized formatter type \"%v\"", formatterType)
}

// Register registers a handler for writer formatter creation
func Register(formatterType string, handler func(map[string]interface{}) (Interface, error)) {
	buildHandlers[formatterType] = handler
}
