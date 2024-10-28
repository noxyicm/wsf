package filter

import (
	"strconv"
	"github.com/noxyicm/wsf/log/event"
)

const (
	// TYPEPriority represents priority filter
	TYPEPriority = "priority"
)

func init() {
	Register(TYPEPriority, NewPriorityFilter)
}

// Priority filters log messages by priority over operator
type Priority struct {
	priority int
	operator string
}

// Accept returns true to accept the message, false to block it
func (p *Priority) Accept(e *event.Event) bool {
	switch p.operator {
	case ">=":
		return e.Priority >= p.priority

	case ">":
		return e.Priority > p.priority

	case "<=":
		return e.Priority <= p.priority

	case "<":
		return e.Priority < p.priority

	case "==":
		return e.Priority == p.priority

	case "!=":
		return e.Priority != p.priority

	default:
		return false
	}
}

// NewPriorityFilter creates priority filter
func NewPriorityFilter(options map[string]interface{}) (Interface, error) {
	p := &Priority{
		operator: "<=",
	}
	if v, ok := options["priority"]; ok {
		switch v.(type) {
		case int:
			p.priority = v.(int)
		case int8:
			p.priority = int(v.(int8))
		case int16:
			p.priority = int(v.(int16))
		case int32:
			p.priority = int(v.(int32))
		case int64:
			p.priority = int(v.(int64))
		case float32:
			p.priority = int(v.(float32))
		case float64:
			p.priority = int(v.(float64))
		case string:
			p.priority, _ = strconv.Atoi(v.(string))
		default:
			p.priority = 0
		}
	}

	if v, ok := options["operator"]; ok {
		p.operator = v.(string)
	}

	return p, nil
}
