package writer

import "wsf/log/event"

const (
	// TYPENull represents null writer
	TYPENull = "null"
)

func init() {
	Register(TYPENull, NewNullWriter)
}

// Null ignors log events
type Null struct {
	writer
}

// Write writes message to log
func (w *Null) Write(e *event.Event) error {
	return nil
}

// Shutdown performs activites such as closing open resources
func (w *Null) Shutdown() {
}

// NewNullWriter creates mock writer
func NewNullWriter(options *Config) (Interface, error) {
	w := &Null{}
	return w, nil
}
