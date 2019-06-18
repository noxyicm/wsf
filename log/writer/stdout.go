package writer

import (
	"fmt"
	"wsf/log/event"
	"wsf/log/filter"
	"wsf/log/formatter"
)

const (
	// TYPEStdout represents stdout writer
	TYPEStdout = "stdout"
)

func init() {
	Register(TYPEStdout, NewStdoutWriter)
}

// Stdout writes log events to stdout
type Stdout struct {
	writer
}

// Write writes message to log
func (w *Stdout) Write(e *event.Event) error {
	for _, filter := range w.filters {
		if !filter.Accept(e) {
			return nil
		}
	}

	message, err := w.formatter.Format(e)
	if err == nil {
		fmt.Print(message)
	} else {
		fmt.Println(err.Error())
	}

	return nil
}

// Shutdown performs activites such as closing open resources
func (w *Stdout) Shutdown() {
}

// NewStdoutWriter creates mock writer
func NewStdoutWriter(options *Config) (Interface, error) {
	w := &Stdout{}

	frt, err := formatter.NewFormatter(options.Formatter)
	if err != nil {
		return nil, err
	}
	w.formatter = frt

	for _, filterParams := range options.Filters {
		flt, err := filter.NewFilter(filterParams)
		if err != nil {
			return nil, err
		}

		w.filters = append(w.filters, flt)
	}

	return w, nil
}
