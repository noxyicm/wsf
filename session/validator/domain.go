package validator

import (
	"wsf/config"
	"wsf/errors"
)

// Public constants
const (
	// TYPESessionValidatorDomain type of validator
	TYPESessionValidatorDomain = "sessionvalidatordomain"
)

func init() {
	Register(TYPESessionValidatorDomain, NewDomain)
}

// Domain validates a session domain
type Domain struct {
	name string
}

// Name returns validator name
func (v *Domain) Name() string {
	return v.name
}

// Setup the validator
func (v *Domain) Setup() error {
	return nil
}

// Valid returns error if validation fails
func (v *Domain) Valid(params map[string]interface{}) error {
	if interfacedValue, ok := params[v.Name()]; ok {
		if typedValue, ok := interfacedValue.(string); ok {
			if typedValue == config.App.GetString("application.Domain") {
				return nil
			}
		}
	}

	return errors.New("Wrong domain")
}

// NewDomain creates a new validator of type Domain
func NewDomain(options *Config) (Interface, error) {
	return &Domain{
		name: "domain",
	}, nil
}
