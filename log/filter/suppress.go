package filter

import "wsf/log/event"

const (
	// TYPESuppress represents suppress filter
	TYPESuppress = "suppress"
)

func init() {
	Register(TYPESuppress, NewSuppressFilter)
}

// Suppress suppresses log messages
type Suppress struct {
	accept bool
}

// Accept returns true to accept the message, false to block it
func (s *Suppress) Accept(e *event.Event) bool {
	return s.accept
}

// NewSuppressFilter creates suppress filter
func NewSuppressFilter(options map[string]interface{}) (Interface, error) {
	s := &Suppress{}
	if v, ok := options["accept"]; ok {
		s.accept = v.(bool)
	}

	return s, nil
}
