package placeholder

import (
	"html"
	"github.com/noxyicm/wsf/view/helper/placeholder/container"
)

// Standalone is a base struct for targetted placeholder helpers
type Standalone struct {
	RegKey     string
	Registry   *Registry
	container  container.Interface
	autoEscape bool
}

// SetContainer sets a container for this placeholder
func (sto *Standalone) SetContainer(ctr container.Interface) error {
	sto.container = ctr
	return nil
}

// Container returns container assosiated with placeholder
func (sto *Standalone) Container() container.Interface {
	return sto.container
}

// SetAutoEscape sets whether or not auto escaping should be used
func (sto *Standalone) SetAutoEscape(escape bool) error {
	sto.autoEscape = escape
	return nil
}

// AutoEscape returns whether or not auto escaping should be used
func (sto *Standalone) AutoEscape() bool {
	return sto.autoEscape
}

// Escape escapes a string
func (sto *Standalone) Escape(escaping string) string {
	return html.EscapeString(escaping)
}

// NewStandaloneContainer creates a new Standalone container
func NewStandaloneContainer(key string) (*Standalone, error) {
	return &Standalone{
		RegKey:     key,
		Registry:   GetRegistry(),
		container:  GetRegistry().GetContainer(key),
		autoEscape: true,
	}, nil
}
