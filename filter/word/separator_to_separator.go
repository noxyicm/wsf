package word

import (
	"reflect"
	"regexp"
	"github.com/noxyicm/wsf/filter"
)

func init() {
	filter.Register("Word_SeparatorToSeparator", reflect.TypeOf((*SeparatorToSeparator)(nil)).Elem())
}

// SeparatorToSeparator inflector
type SeparatorToSeparator struct {
	filter.RegexpReplace
	searchSeparator      string
	replacementSeparator string
}

// Filter applyes filter
func (s *SeparatorToSeparator) Filter(value interface{}) (interface{}, error) {
	return s.RegexpReplace.Filter(value)
}

// Defaults sets default properties
func (s *SeparatorToSeparator) Defaults() error {
	err := s.RegexpReplace.Defaults()
	if err != nil {
		return err
	}

	s.searchSeparator = " "
	s.SetMatchPattern(regexp.QuoteMeta(s.searchSeparator))
	s.SetReplacement(s.replacementSeparator)

	return nil
}

// SetSearchSeparator sets search separator string
func (s *SeparatorToSeparator) SetSearchSeparator(sep string) error {
	s.searchSeparator = sep
	s.SetMatchPattern(regexp.QuoteMeta(s.searchSeparator))
	return nil
}

// SetReplacementSeparator sets replacement separator string
func (s *SeparatorToSeparator) SetReplacementSeparator(sep string) error {
	s.replacementSeparator = sep
	s.SetReplacement(s.replacementSeparator)
	return nil
}

// NewSeparatorToSeparator creates new separator to separator inflector
func NewSeparatorToSeparator(fromsep string, tosep string) (filter.Interface, error) {
	u := &SeparatorToSeparator{}
	err := u.Defaults()
	if err != nil {
		return nil, err
	}

	u.SetSearchSeparator(fromsep)
	u.SetReplacementSeparator(tosep)
	return u, nil
}
