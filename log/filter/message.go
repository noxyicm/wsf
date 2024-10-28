package filter

import (
	"regexp"
	"github.com/noxyicm/wsf/log/event"
)

const (
	// TYPEMessage represents message filter
	TYPEMessage = "message"
)

func init() {
	Register(TYPEMessage, NewMessageFilter)
}

// Message filters log messages by regexp
type Message struct {
	reg *regexp.Regexp
}

// Accept returns true to accept the message, false to block it
func (m *Message) Accept(e *event.Event) bool {
	return m.reg.MatchString(e.Message)
}

// NewMessageFilter creates message filter
func NewMessageFilter(options map[string]interface{}) (mi Interface, err error) {
	m := &Message{}
	if v, ok := options["regexp"]; ok {
		m.reg, err = regexp.Compile(v.(string))
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
