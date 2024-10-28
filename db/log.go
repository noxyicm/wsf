package db

import (
	"wsf/config"
	"wsf/context"
	"wsf/errors"
)

// Public constants
const (
	OperationAdd    = "add"
	OperationEdit   = "edit"
	OperationDelete = "delete"
)

var (
	buildLogHandlers = map[string]func(config.Config) (Log, error){}
)

type Log interface {
	Enable() Log
	Disable() Log
	IsEnabled() bool
	SetAdvanced(value bool) Log
	IsAdvanced() bool
	SetExcludes(tables []string) Log
	AddExclude(table string) Log
	Excludes() []string
	IsExcluded(table string) bool
	IsWritable(table string) bool
	Write(ctx context.Context, operation string, table string, id int, data map[string]interface{}) bool
}

// NewLog creates a new log from given type and options
func NewLog(logType string, options config.Config) (lg Log, err error) {
	if f, ok := buildLogHandlers[logType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized database log type \"%v\"", logType)
}

// RegisterLogger registers a handler for database logger creation
func RegisterLogger(logType string, handler func(config.Config) (Log, error)) {
	buildLogHandlers[logType] = handler
}
