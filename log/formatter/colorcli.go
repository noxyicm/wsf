package formatter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log/event"

	"github.com/mgutz/ansi"
)

const (
	// TYPEColorized represents colorized formatter
	TYPEColorized = "colorized"

	// ColorizedFormat is a default format for simple formatter
	ColorizedFormat = "<white+hd>[#timestamp#]</reset> <%s>#priorityName# (#priority#): #message#</reset>\n"

	// EMERG represents emergency event
	EMERG = 0 // Emergency: system is unusable
	// ALERT represents alert event
	ALERT = 1 // Alert: action must be taken immediately
	// CRIT represents critical error event
	CRIT = 2 // Critical: critical conditions
	// ERR represents general error event
	ERR = 3 // Error: error conditions
	// WARN represents warning event
	WARN = 4 // Warning: warning conditions
	// NOTICE represents notice event
	NOTICE = 5 // Notice: normal but significant condition
	// INFO represents informational event
	INFO = 6 // Informational: informational messages
	// DEBUG represents debug event
	DEBUG = 7 // Debug: debug messages
)

func init() {
	Register(TYPEColorized, NewColorizedFormatter)
}

// ColorCli is a colorizing formatter
type ColorCli struct {
	format string
	reg    *regexp.Regexp
}

// Format formats data into a single line to be written by the writer
func (f *ColorCli) Format(e *event.Event) (string, error) {
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

	switch e.Priority {
	case INFO:
		out = f.Sprintf(out, "green")
	case DEBUG, NOTICE:
		out = f.Sprintf(out, "white")
	case WARN:
		out = f.Sprintf(out, "yellow")
	case ERR, CRIT, ALERT, EMERG:
		out = f.Sprintf(out, "red")
	}

	return out, nil
}

// Sprintf works identically to fmt.Sprintf
func (f *ColorCli) Sprintf(format string, args ...interface{}) string {
	format = fmt.Sprintf(format, args...)
	format = f.reg.ReplaceAllStringFunc(format, func(s string) string {
		return ansi.ColorCode(strings.Trim(s, "<>/"))
	})

	return format
}

// NewColorizedFormatter creates colorizing formatter
func NewColorizedFormatter(options map[string]interface{}) (Interface, error) {
	reg, _ := regexp.Compile(`<([^>]+)>`)
	f := &ColorCli{
		format: ColorizedFormat,
		reg:    reg,
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
