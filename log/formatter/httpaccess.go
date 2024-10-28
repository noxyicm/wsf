package formatter

import (
	"strings"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log/event"
)

const (
	// TYPEHTTPAccess represents http access formatter
	TYPEHTTPAccess = "httpaccess"

	// HTTPAccessFormat is a default format for http access formatter
	HTTPAccessFormat = "%client% - %user% [%timestamp%] \"%request%\" %statusCode% %bytes% \"%referer%\" \"%useragent%\"\n"
)

func init() {
	Register(TYPEHTTPAccess, NewHTTPAccessFormatter)
}

// HTTPAccess filters log messages by priority over operator
type HTTPAccess struct {
	format string
}

// Format formats data into a single line to be written by the writer
func (f *HTTPAccess) Format(e *event.Event) (string, error) {
	out := f.format
	out = strings.Replace(out, "%timestamp%", e.Timestamp, 1)

	for key, value := range e.Info {
		out = strings.Replace(out, "%"+key+"%", value, 1)
	}

	return out, nil
}

// NewHTTPAccessFormatter creates http access formatter
func NewHTTPAccessFormatter(options map[string]interface{}) (Interface, error) {
	f := &HTTPAccess{
		format: HTTPAccessFormat,
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
