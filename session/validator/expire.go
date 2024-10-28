package validator

import (
	"time"
	"github.com/noxyicm/wsf/errors"
)

// Public constants
const (
	// TYPESessionValidatorExpire type of validator
	TYPESessionValidatorExpire = "sessionvalidatorexpire"
)

func init() {
	Register(TYPESessionValidatorExpire, NewExpire)
}

// Expire validates a session expiration time
type Expire struct {
	name string
}

// Name returns validator name
func (v *Expire) Name() string {
	return v.name
}

// Setup the validator
func (v *Expire) Setup() error {
	return nil
}

// Valid returns error if validation fails
func (v *Expire) Valid(params map[string]interface{}) error {
	if interfacedValue, ok := params[v.Name()]; ok {
		if typedValue, ok := interfacedValue.(int64); ok {
			if typedValue < time.Now().Unix() {
				return nil
			}
		}
	}

	return errors.New("Session expired")
}

// NewExpire creates a new validator of type Expire
func NewExpire(options *Config) (Interface, error) {
	return &Expire{
		name: "expire",
	}, nil
}
