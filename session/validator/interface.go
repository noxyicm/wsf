package validator

import (
	"wsf/config"
	"wsf/errors"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

// Interface is a session validator interface
type Interface interface {
	Name() string
	Setup() error
	Valid(params map[string]interface{}) error
}

// NewValidator creates a new validator
func NewValidator(validatorType string, options config.Config) (Interface, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[validatorType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized validator type \"%v\"", validatorType)
}

// NewValidatorFromConfig creates a new validator from Config
func NewValidatorFromConfig(options *Config) (Interface, error) {
	if f, ok := buildHandlers[options.Type]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized validator type \"%v\"", options.Type)
}

// Register registers a handler for validator creation
func Register(validatorType string, handler func(*Config) (Interface, error)) {
	if _, ok := buildHandlers[validatorType]; ok {
		panic("[Session] Validator of type '" + validatorType + "' is already registered")
	}

	buildHandlers[validatorType] = handler
}
