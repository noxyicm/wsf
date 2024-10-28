package word

import (
	"reflect"
	"github.com/noxyicm/wsf/filter"
)

func init() {
	filter.Register("Word_Separator", reflect.TypeOf((*Separator)(nil)).Elem())
}

// Separator inflector
type Separator struct {
	filter.RegexpReplace
	separator string
}

// Filter applyes filter
func (s *Separator) Filter(value interface{}) (interface{}, error) {
	return s.RegexpReplace.Filter(value)
}

// Defaults sets default properties
func (s *Separator) Defaults() error {
	return s.RegexpReplace.Defaults()
}

// SetSeparator sets the separator
func (s *Separator) SetSeparator(separator string) error {
	s.separator = separator
	return nil
}

// NewSeparator creates new separator inflector
func NewSeparator(sep string) (filter.Interface, error) {
	sprt := &Separator{
		separator: sep,
	}
	err := sprt.Defaults()
	if err != nil {
		return nil, err
	}

	return sprt, nil
}
