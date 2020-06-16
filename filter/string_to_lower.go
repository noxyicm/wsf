package filter

import (
	"reflect"
	"strings"

	"wsf/errors"
)

func init() {
	Register("StringToLower", reflect.TypeOf((*StringToLower)(nil)).Elem())
}

// StringToLower filter
type StringToLower struct {
}

// Filter applyes filter
func (s *StringToLower) Filter(value interface{}) (interface{}, error) {
	if v, ok := value.(string); ok {
		return strings.ToLower(v), nil
	}

	return value, errors.Errorf("Value %v is not a string", value)
}

// Defaults sets default properties
func (s *StringToLower) Defaults() error {
	return nil
}

// NewStringToLower creates new string to lower inflector
func NewStringToLower() (Interface, error) {
	s := &StringToLower{}
	return s, nil
}
