package word

import (
	"reflect"
	"github.com/noxyicm/wsf/filter"
)

func init() {
	filter.Register("Word_UnderscoreToSeparator", reflect.TypeOf((*UnderscoreToSeparator)(nil)).Elem())
}

// UnderscoreToSeparator filter
type UnderscoreToSeparator struct {
	SeparatorToSeparator
}

// Filter applyes filter
func (u *UnderscoreToSeparator) Filter(value interface{}) (interface{}, error) {
	return u.SeparatorToSeparator.Filter(value)
}

// Defaults sets default properties
func (u *UnderscoreToSeparator) Defaults() error {
	err := u.SeparatorToSeparator.Defaults()
	if err != nil {
		return err
	}

	u.SetSearchSeparator("_")
	return nil
}

// NewUnderscoreToSeparator creates new underscore to separator inflector
func NewUnderscoreToSeparator(sep string) (filter.Interface, error) {
	u := &UnderscoreToSeparator{}
	u.SetReplacementSeparator(sep)
	err := u.Defaults()
	if err != nil {
		return nil, err
	}

	return u, nil
}
