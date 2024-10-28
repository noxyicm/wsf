package writer

import (
	"fmt"
	"github.com/noxyicm/wsf/log/event"
	"github.com/noxyicm/wsf/log/filter"
	"github.com/noxyicm/wsf/log/formatter"
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
	if !w.Enable {
		return nil
	}

	for _, filter := range w.Filters {
		if !filter.Accept(e) {
			return nil
		}
	}

	message, err := w.Formatter.Format(e)
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
	w.Enable = options.Enable
	frt, err := formatter.NewFormatter(options.Formatter)
	if err != nil {
		return nil, err
	}
	w.Formatter = frt

	for _, filterParams := range options.Filters {
		flt, err := filter.NewFilter(filterParams)
		if err != nil {
			return nil, err
		}

		w.Filters = append(w.Filters, flt)
	}

	return w, nil
}
