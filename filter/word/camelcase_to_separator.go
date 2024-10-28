package word

import (
	"reflect"
	"github.com/noxyicm/wsf/filter"
)

func init() {
	filter.Register("Word_CamelCaseToSeparator", reflect.TypeOf((*CamelCaseToSeparator)(nil)).Elem())
}

// CamelCaseToSeparator inflector
type CamelCaseToSeparator struct {
	Separator
}

// Filter applyes filter
func (c *CamelCaseToSeparator) Filter(value interface{}) (interface{}, error) {
	return c.Separator.Filter(value)
}

// Defaults sets default properties
func (c *CamelCaseToSeparator) Defaults() error {
	err := c.Separator.Defaults()
	if err != nil {
		return err
	}

	if c.IsUnicodeSupportEnabled() {
		err := c.SetMatchPatterns([]string{`(\p{Lu})(\p{Lu}\p{Ll})`, `(\p{Ll}|\p{Nd}))(\p{Lu})`})
		if err != nil {
			return err
		}
		c.SetReplacements([]string{"$1" + c.separator + "$2", "$1" + c.separator + "$2"})
	} else {
		err := c.SetMatchPatterns([]string{`([A-Z])([A-Z]+)([A-Z][A-z])`, `([a-z0-9])([A-Z])`})
		if err != nil {
			return err
		}
		c.SetReplacements([]string{"$1" + c.separator + "$2", "$1" + c.separator + "$2"})
	}

	return nil
}

// NewCamelCaseToSeparator creates new camelcase to separator inflector
func NewCamelCaseToSeparator(sep string) (filter.Interface, error) {
	c := &CamelCaseToSeparator{}
	err := c.Defaults()
	if err != nil {
		return nil, err
	}

	return c, nil
}
