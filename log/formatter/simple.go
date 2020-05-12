package formatter

import (
	"strconv"
	"strings"
	"wsf/errors"
	"wsf/log/event"
)

const (
	// TYPESimple represents simple formatter
	TYPESimple = "simple"

	// SimpleFormat is a default format for simple formatter
	SimpleFormat = "[#timestamp#] #priorityName# (#priority#): #message#\n"
)

func init() {
	Register(TYPESimple, NewSimpleFormatter)
}

// Simple filters log messages by priority over operator
type Simple struct {
	format string
}

// Format formats data into a single line to be written by the writer
func (f *Simple) Format(e *event.Event) (string, error) {
	out := f.format

	out = strings.Replace(out, "#timestamp#", e.Timestamp, 1)
	out = strings.Replace(out, "#priorityName#", e.PriorityName, 1)
	out = strings.Replace(out, "#priority#", strconv.Itoa(e.Priority), 1)
	out = strings.Replace(out, "#message#", e.Message, 1)

	for key, value := range e.Info {
		out = strings.Replace(out, "#"+key+"#", value, 1)
	}

	m := rest.FindAllString(out, -1)
	for i := range m {
		out = strings.Replace(out, m[i], "", 1)
	}

	return out, nil
}

// NewSimpleFormatter creates simple formatter
func NewSimpleFormatter(options map[string]interface{}) (Interface, error) {
	f := &Simple{
		format: SimpleFormat,
	}

	if v, ok := options["format"]; ok {
		if v, ok := v.(string); ok {
			f.format = v
		} else {
			return nil, errors.New("Format must be string")
		}
	}

	return f, nil
}
