package resource

import (
	"wsf/config"
	"wsf/log"
)

// TYPELogger id of resource
const TYPELogger = "log"

func init() {
	Register(TYPELogger, NewLoggerResource)
}

// NewLoggerResource creates a new resource of type Log
func NewLoggerResource(cfg config.Config) (Interface, error) {
	return log.NewLog(cfg)
}
