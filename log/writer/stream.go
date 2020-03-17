package writer

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"wsf/config"
	"wsf/errors"
	"wsf/log/event"
	"wsf/log/filter"
	"wsf/log/formatter"
)

const (
	// TYPEStream represents null writer
	TYPEStream = "stream"
)

func init() {
	Register(TYPEStream, NewStreamWriter)
}

// Stream writes log events to a stream
type Stream struct {
	writer
	stream *os.File
	mode   int
}

// Write writes message to log
func (w *Stream) Write(e *event.Event) error {
	for _, filter := range w.filters {
		if !filter.Accept(e) {
			return nil
		}
	}

	message, err := w.formatter.Format(e)
	if err == nil {
		if _, err := w.stream.Write([]byte(message)); err != nil {
			return err
		}
	}

	return nil
}

// Shutdown performs activites such as closing open resources
func (w *Stream) Shutdown() {
	if err := w.stream.Close(); err != nil {
		log.Fatal(err)
	}
}

// NewStreamWriter creates stream writer
func NewStreamWriter(options *Config) (Interface, error) {
	w := &Stream{}
	if v, ok := options.Params["mode"]; ok {
		w.mode = v.(int)
	} else {
		w.mode = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	}

	if v, ok := options.Params["stream"]; ok {
		path := filepath.Join(config.AppRootPath, filepath.FromSlash(v.(string)))
		if p, err := filepath.EvalSymlinks(path); err == nil {
			path = p
		}

		file, err := os.OpenFile(path, w.mode, 0644)
		if err != nil {
			file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				parts := strings.Split(path, "/")
				dir := strings.Join(parts[:len(parts)-1], "/")
				if err := os.MkdirAll(dir, 0775); err != nil {
					return nil, errors.Wrap(err, "Failed to create log stream writer")
				}

				file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return nil, errors.Wrap(err, "Failed to create log stream writer")
				}
			}
		}

		w.stream = file
	}

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
