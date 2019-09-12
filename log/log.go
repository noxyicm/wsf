package log

import (
	"fmt"
	"strings"
	"time"
	"wsf/config"
	"wsf/errors"
	"wsf/log/event"
	"wsf/log/filter"
	"wsf/log/writer"
)

// Log messages level
const (
	EMERG  = 0 // Emergency: system is unusable
	ALERT  = 1 // Alert: action must be taken immediately
	CRIT   = 2 // Critical: critical conditions
	ERR    = 3 // Error: error conditions
	WARN   = 4 // Warning: warning conditions
	NOTICE = 5 // Notice: normal but significant condition
	INFO   = 6 // Informational: informational messages
	DEBUG  = 7 // Debug: debug messages
)

var (
	priorities = make(map[int]string)
	lg         *Log
)

func init() {
	priorities[EMERG] = "EMERGENCY"
	priorities[ALERT] = "ALERT"
	priorities[CRIT] = "CRITICAL"
	priorities[ERR] = "ERROR"
	priorities[WARN] = "WARNING"
	priorities[NOTICE] = "NOTICE"
	priorities[INFO] = "INFO"
	priorities[DEBUG] = "DEBUG"
}

// Log writes log
type Log struct {
	options    *Config
	enable     bool
	priorities map[int]string
	writers    []writer.Interface
	filters    []filter.Interface
	extras     map[string]string

	timestampFormat string
}

// Priority returns resource initialization priority
func (l *Log) Priority() int {
	return l.options.Priority
}

// AddPriority add a custom priority
func (l *Log) AddPriority(name string, priority int) error {
	name = strings.ToUpper(name)

	if _, ok := l.priorities[priority]; ok {
		return errors.New("Existing priorities cannot be overwritten")
	}

	l.priorities[priority] = name
	return nil
}

// AddFilter addd a filter that will be applied before all log writers
// Before a message will be received by any of the writers, it
// must be accepted by all filters added with this method
func (l *Log) AddFilter(options interface{}) (err error) {
	var newfilter filter.Interface
	switch options.(type) {
	case map[string]interface{}:
		newfilter, err = filter.NewFilter(options.(map[string]interface{}))
		if err != nil {
			return err
		}

	case int:
		newfilter, err = filter.NewFilter(map[string]interface{}{
			"type":     "priority",
			"priority": options.(int),
			"operator": "<=",
		})
		if err != nil {
			return err
		}

	case filter.Interface:
		newfilter = options.(filter.Interface)

	default:
		return errors.New("Invalid filter provided")
	}

	l.filters = append(l.filters, newfilter)
	return nil
}

// AddWriter adds a writer. A writer is responsible for taking a log
// message and writing it out to storage
func (l *Log) AddWriter(options interface{}) (err error) {
	var newwriter writer.Interface
	switch options.(type) {
	case *writer.Config:
		newwriter, err = writer.NewWriter(options.(*writer.Config))
		if err != nil {
			return err
		}

	case writer.Interface:
		newwriter = options.(writer.Interface)

	default:
		return errors.New("Writer must be an instance of writer.Interface or you should pass a writer configuration")
	}

	l.writers = append(l.writers, newwriter)
	return nil
}

// SetTimestampFormat sets timestamp format for log entries
func (l *Log) SetTimestampFormat(format string) {
	l.timestampFormat = format
}

// Log a message at a priority
func (l *Log) Log(message string, priority int, extras map[string]string) error {
	if !l.enable {
		return nil
	}

	if _, ok := l.priorities[priority]; !ok {
		return errors.New("Bad log priority")
	}

	event, err := l.packEvent(message, priority)
	if err != nil {
		return errors.Wrap(err, "[Log] Pack event error")
	}

	if len(extras) > 0 {
		for key, value := range extras {
			event.Info[key] = value
		}
	}

	for _, filter := range l.filters {
		if !filter.Accept(event) {
			return nil
		}
	}

	for _, writer := range l.writers {
		writer.Write(event)
	}

	return nil
}

// Logf logs a formated message at a priority
func (l *Log) Logf(message string, priority int, extras map[string]string, f ...interface{}) error {
	message = fmt.Sprintf(message, f...)
	return l.Log(message, priority, extras)
}

// Debug logs message at a debug priority
func (l *Log) Debug(message string, extras map[string]string) error {
	return l.Log(message, DEBUG, extras)
}

// Debugf logs formated message at a debug priority
func (l *Log) Debugf(message string, extras map[string]string, f ...interface{}) error {
	return l.Logf(message, DEBUG, extras, f...)
}

// Info logs message at an info priority
func (l *Log) Info(message string, extras map[string]string) error {
	return l.Log(message, INFO, extras)
}

// Infof logs formated message at an info priority
func (l *Log) Infof(message string, extras map[string]string, f ...interface{}) error {
	return l.Logf(message, INFO, extras, f...)
}

// Notice logs message at a notice priority
func (l *Log) Notice(message string, extras map[string]string) error {
	return l.Log(message, NOTICE, extras)
}

// Noticef logs formated message at a notice priority
func (l *Log) Noticef(message string, extras map[string]string, f ...interface{}) error {
	return l.Logf(message, NOTICE, extras, f...)
}

// Warning logs message at a warning priority
func (l *Log) Warning(message string, extras map[string]string) error {
	return l.Log(message, WARN, extras)
}

// Warningf logs formated message at a warning priority
func (l *Log) Warningf(message string, extras map[string]string, f ...interface{}) error {
	return l.Logf(message, WARN, extras, f...)
}

// Error logs message at an error priority
func (l *Log) Error(message string, extras map[string]string) error {
	return l.Log(message, ERR, extras)
}

// Errorf logs formated message at an error priority
func (l *Log) Errorf(message string, extras map[string]string, f ...interface{}) error {
	return l.Logf(message, ERR, extras, f...)
}

// Critical logs message at a critical priority
func (l *Log) Critical(message string, extras map[string]string) error {
	return l.Log(message, CRIT, extras)
}

// Criticalf logs formated message at a critical priority
func (l *Log) Criticalf(message string, extras map[string]string, f ...interface{}) error {
	return l.Logf(message, CRIT, extras, f...)
}

// Alert logs message at an alert priority
func (l *Log) Alert(message string, extras map[string]string) error {
	return l.Log(message, ALERT, extras)
}

// Alertf logs formated message at an alert priority
func (l *Log) Alertf(message string, extras map[string]string, f ...interface{}) error {
	return l.Logf(message, ALERT, extras, f...)
}

// Emergency logs message at an emergency priority
func (l *Log) Emergency(message string, extras map[string]string) error {
	return l.Log(message, EMERG, extras)
}

// Emergencyf logs formated message at an emergency priority
func (l *Log) Emergencyf(message string, extras map[string]string, f ...interface{}) error {
	return l.Logf(message, EMERG, extras, f...)
}

// Destroy shuts down all writers
func (l *Log) Destroy() {
	for _, writer := range l.writers {
		writer.Shutdown()
	}
}

func (l *Log) packEvent(message string, priority int) (*event.Event, error) {
	if _, ok := l.priorities[priority]; !ok {
		return nil, errors.New("Bad log priority")
	}

	e := &event.Event{
		Timestamp:    time.Now().Format(l.timestampFormat),
		Message:      message,
		Priority:     priority,
		PriorityName: l.priorities[priority],
		Info:         make(map[string]string),
	}
	return e, nil
}

// NewLog creates new logger
func NewLog(options config.Config) (*Log, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	l := &Log{
		options:         cfg,
		enable:          cfg.Enable,
		priorities:      priorities,
		timestampFormat: cfg.TimestampFormat,
	}

	l.extras = cfg.Extras
	for _, filtersParams := range cfg.Filters {
		newfilter, err := filter.NewFilter(filtersParams)
		if err != nil {
			return nil, err
		}

		l.filters = append(l.filters, newfilter)
	}

	for _, writerParams := range cfg.Writers {
		newwriter, err := writer.NewWriter(writerParams)
		if err != nil {
			return nil, err
		}

		for _, filterInst := range l.filters {
			if err := newwriter.AddFilter(filterInst); err != nil {
				continue
			}
		}

		l.writers = append(l.writers, newwriter)
	}

	return l, nil
}

// SetInstance sets global log instance
func SetInstance(l *Log) {
	lg = l
}

// Instance returns global log instance
func Instance() *Log {
	return lg
}

// Debug logs message at a debug priority
func Debug(message string, extras map[string]string) error {
	return lg.Log(message, DEBUG, extras)
}

// Debugf logs formated message at a debug priority
func Debugf(message string, extras map[string]string, f ...interface{}) error {
	return lg.Logf(message, DEBUG, extras, f...)
}

// Info logs message at an info priority
func Info(message string, extras map[string]string) error {
	return lg.Log(message, INFO, extras)
}

// Infof logs formated message at an info priority
func Infof(message string, extras map[string]string, f ...interface{}) error {
	return lg.Logf(message, INFO, extras, f...)
}

// Notice logs message at a notice priority
func Notice(message string, extras map[string]string) error {
	return lg.Log(message, NOTICE, extras)
}

// Noticef logs formated message at a notice priority
func Noticef(message string, extras map[string]string, f ...interface{}) error {
	return lg.Logf(message, NOTICE, extras, f...)
}

// Warning logs message at a warning priority
func Warning(message string, extras map[string]string) error {
	return lg.Log(message, WARN, extras)
}

// Warningf logs formated message at a warning priority
func Warningf(message string, extras map[string]string, f ...interface{}) error {
	return lg.Logf(message, WARN, extras, f...)
}

// Error logs message at an error priority
func Error(message string, extras map[string]string) error {
	return lg.Log(message, ERR, extras)
}

// Errorf logs formated message at an error priority
func Errorf(message string, extras map[string]string, f ...interface{}) error {
	return lg.Logf(message, ERR, extras, f...)
}

// Critical logs message at a critical priority
func Critical(message string, extras map[string]string) error {
	return lg.Log(message, CRIT, extras)
}

// Criticalf logs formated message at a critical priority
func Criticalf(message string, extras map[string]string, f ...interface{}) error {
	return lg.Logf(message, CRIT, extras, f...)
}

// Alert logs message at an alert priority
func Alert(message string, extras map[string]string) error {
	return lg.Log(message, ALERT, extras)
}

// Alertf logs formated message at an alert priority
func Alertf(message string, extras map[string]string, f ...interface{}) error {
	return lg.Logf(message, ALERT, extras, f...)
}

// Emergency logs message at an emergency priority
func Emergency(message string, extras map[string]string) error {
	return lg.Log(message, EMERG, extras)
}

// Emergencyf logs formated message at an emergency priority
func Emergencyf(message string, extras map[string]string, f ...interface{}) error {
	return lg.Logf(message, EMERG, extras, f...)
}
