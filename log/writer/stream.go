package writer

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log/event"
	"github.com/noxyicm/wsf/log/filter"
	"github.com/noxyicm/wsf/log/formatter"
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
	mur    sync.RWMutex
}

// Write writes message to log
func (w *Stream) Write(e *event.Event) error {
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
		w.mur.Lock()
		if _, err := w.stream.Write([]byte(message)); err != nil {
			w.mur.Unlock()
			return err
		}
		w.mur.Unlock()
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
	w.Enable = options.Enable
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
