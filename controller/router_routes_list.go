package controller

import (
	"wsf/errors"
)

// RoutesList is referenced stack of routes
type RoutesList struct {
	stack    []RouteInterface
	refs     map[string]int
	backrefs map[int]string
}

// Set values as stack
func (s *RoutesList) Set(key string, value RouteInterface) error {
	if k, ok := s.refs[key]; ok {
		s.stack[k] = value
		return nil
	}

	return errors.Errorf("Key '%s' does not exists", key)
}

// Append value to stack
func (s *RoutesList) Append(key string, value RouteInterface) error {
	if _, ok := s.refs[key]; ok {
		return errors.Errorf("Key '%s' already exists. Use Set() instead", key)
	}

	s.stack = append(s.stack, value)
	s.refs[key] = len(s.stack) - 1
	s.backrefs[s.refs[key]] = key
	return nil
}

// Prepend value to stack
func (s *RoutesList) Prepend(key string, value RouteInterface) error {
	if _, ok := s.refs[key]; ok {
		return errors.Errorf("Key '%s' already exists. Use Set() instead", key)
	}

	s.pushup(0)

	s.stack = append([]RouteInterface{value}, s.stack...)
	s.refs[key] = 0
	s.backrefs[0] = key
	return nil
}

// InsertBefore inserts value before parent
func (s *RoutesList) InsertBefore(parent string, key string, value RouteInterface) error {
	if _, ok := s.refs[key]; ok {
		return errors.Errorf("Key '%s' already exists. Use Set() instead", key)
	}

	if k, ok := s.refs[parent]; ok {
		s.pushup(k)

		if k == 0 {
			s.stack = append([]RouteInterface{value}, s.stack...)
		} else {
			s.stack = append(s.stack[:k], append([]RouteInterface{value}, s.stack[k:]...)...)
		}

		s.refs[key] = k
		s.backrefs[k] = key
	} else {
		return errors.Errorf("Key '%s' does not exists", parent)
	}

	return nil
}

// InsertAfter inserts value after parent
func (s *RoutesList) InsertAfter(parent string, key string, value RouteInterface) error {
	if _, ok := s.refs[key]; ok {
		return errors.Errorf("Key '%s' already exists. Use Set() instead", key)
	}

	if k, ok := s.refs[parent]; ok {
		k = k + 1
		s.pushup(k)

		s.stack = append(s.stack[:k], append([]RouteInterface{value}, s.stack[k:]...)...)
		s.refs[key] = k
		s.backrefs[k] = key
	} else {
		return errors.Errorf("Key '%s' does not exists", parent)
	}

	return nil
}

// Pop value from the top of the stack
func (s *RoutesList) Pop() RouteInterface {
	k := len(s.stack) - 1
	value := s.stack[k]

	s.Unset(s.backrefs[k])
	return value
}

// Unset deletes value from stack by key
func (s *RoutesList) Unset(key string) error {
	if k, ok := s.refs[key]; ok {
		s.stack = append(s.stack[:k], s.stack[k:]...)
		delete(s.backrefs, k)
		delete(s.refs, key)
		s.pushdown(k)
	} else {
		return errors.Errorf("Key '%s' does not exists", key)
	}

	return nil
}

// Stack returns all values
func (s *RoutesList) Stack() []RouteInterface {
	return s.stack
}

// Value returns value by key
func (s *RoutesList) Value(key string) RouteInterface {
	if v, ok := s.refs[key]; ok {
		return s.stack[v]
	}

	return nil
}

// Key returns key for provided value
func (s *RoutesList) Key(value RouteInterface) (string, bool) {
	for i, v := range s.stack {
		if v == value {
			return s.backrefs[i], true
		}
	}

	return "", false
}

// IKey returns key for provided integer stack index
func (s *RoutesList) IKey(key int) (string, bool) {
	if v, ok := s.backrefs[key]; ok {
		return v, true
	}

	return "", false
}

// Map retruns stack as map
func (s *RoutesList) Map() map[string]RouteInterface {
	m := make(map[string]RouteInterface)
	for k, v := range s.stack {
		m[s.backrefs[k]] = v
	}

	return m
}

// Contains returns true if stack contains a value
func (s *RoutesList) Contains(value RouteInterface) bool {
	for _, v := range s.stack {
		if v == value {
			return true
		}
	}

	return false
}

// Has returns true if stack contains a key
func (s *RoutesList) Has(key string) bool {
	if _, ok := s.refs[key]; ok {
		return true
	}

	return false
}

// Clear clears the stack
func (s *RoutesList) Clear() error {
	s.stack = make([]RouteInterface, 0)
	s.refs = make(map[string]int)
	s.backrefs = make(map[int]string)
	return nil
}

// Populate the stack with provided values
func (s *RoutesList) Populate(data map[string]RouteInterface) {
	for k, v := range data {
		s.Append(k, v)
	}
}

// Reverce sort list in reverce order
func (s *RoutesList) Reverce() {

}

func (s *RoutesList) pushup(key int) {
	newbackrefs := make(map[int]string)
	for k := range s.stack {
		if k >= key {
			newbackrefs[k+1] = s.backrefs[k]
			s.refs[s.backrefs[k]] = k + 1
		}
	}
	s.backrefs = newbackrefs
}

func (s *RoutesList) pushdown(key int) {
	newbackrefs := make(map[int]string)
	for k := range s.stack {
		if k >= key {
			newbackrefs[k-1] = s.backrefs[k]
			s.refs[s.backrefs[k]] = k - 1
		}
	}
	s.backrefs = newbackrefs
}

// NewRoutesList creates a new referenced stack
func NewRoutesList() *RoutesList {
	s := &RoutesList{
		refs:     make(map[string]int),
		stack:    make([]RouteInterface, 0),
		backrefs: make(map[int]string),
	}

	return s
}

// ReverseRoutesList reverce order of RouteList
func ReverseRoutesList(list *RoutesList) *RoutesList {
	rl := NewRoutesList()
	slice := list.Stack()

	s := make([]RouteInterface, len(slice))
	copy(s, slice)
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	for _, r := range s {
		if v, ok := list.Key(r); ok {
			rl.Append(v, r)
		}
	}

	return rl
}
