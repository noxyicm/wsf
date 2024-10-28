package word

import (
	"reflect"
	"github.com/noxyicm/wsf/filter"
)

func init() {
	filter.Register("Word_CamelCaseToDash", reflect.TypeOf((*CamelCaseToDash)(nil)).Elem())
}

// CamelCaseToDash inflector
type CamelCaseToDash struct {
	CamelCaseToSeparator
}

// Filter applyes filter
func (c *CamelCaseToDash) Filter(value interface{}) (interface{}, error) {
	return c.CamelCaseToSeparator.Filter(value)
}

// Defaults sets default properties
func (c *CamelCaseToDash) Defaults() error {
	err := c.CamelCaseToSeparator.Defaults()
	if err != nil {
		return err
	}

	c.separator = "-"
	return nil
}

// NewCamelCaseToDash creates new separator inflector
func NewCamelCaseToDash() (filter.Interface, error) {
	c := &CamelCaseToSeparator{}
	err := c.Defaults()
	if err != nil {
		return nil, err
	}

	return c, nil
}
