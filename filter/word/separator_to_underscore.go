package word

import (
	"reflect"

	"github.com/noxyicm/wsf/filter"
)

func init() {
	filter.Register("Word_SeparatorToUnderscore", reflect.TypeOf((*SeparatorToUnderscore)(nil)).Elem())
}

// SeparatorToUnderscore filter
type SeparatorToUnderscore struct {
	SeparatorToSeparator
}

// Filter applyes filter
func (u *SeparatorToUnderscore) Filter(value interface{}) (interface{}, error) {
	return u.SeparatorToSeparator.Filter(value)
}

// Defaults sets default properties
func (u *SeparatorToUnderscore) Defaults() error {
	err := u.SeparatorToSeparator.Defaults()
	if err != nil {
		return err
	}

	u.SetReplacementSeparator("_")
	return nil
}

// NewSeparatorToUnderscore creates new underscore to separator inflector
func NewSeparatorToUnderscore(sep string) (filter.Interface, error) {
	u := &SeparatorToUnderscore{}
	err := u.Defaults()
	if err != nil {
		return nil, err
	}

	u.SetSearchSeparator(sep)
	return u, nil
}
