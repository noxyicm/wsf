package stack

import (
	"github.com/noxyicm/wsf/errors"
)

// Referenced stack
type Referenced struct {
	stack    []interface{}
	refs     map[string]int
	backrefs map[int]string
}

// Set values as stack
func (s *Referenced) Set(key string, value interface{}) error {
	if k, ok := s.refs[key]; ok {
		s.stack[k] = value
		return nil
	}

	return errors.Errorf("Key '%s' does not exists", key)
}

// Append value to stack
func (s *Referenced) Append(key string, value interface{}) error {
	if _, ok := s.refs[key]; ok {
		return errors.Errorf("Key '%s' already exists. Use Set() instead", key)
	}

	s.stack = append(s.stack, value)
	s.refs[key] = len(s.stack) - 1
	s.backrefs[s.refs[key]] = key
	return nil
}

// Prepend value to stack
func (s *Referenced) Prepend(key string, value interface{}) error {
	if _, ok := s.refs[key]; ok {
		return errors.Errorf("Key '%s' already exists. Use Set() instead", key)
	}

	s.pushup(0)

	s.stack = append([]interface{}{value}, s.stack)
	s.refs[key] = 0
	s.backrefs[0] = key
	return nil
}

// InsertBefore inserts value before parent
func (s *Referenced) InsertBefore(parent string, key string, value interface{}) error {
	if _, ok := s.refs[key]; ok {
		return errors.Errorf("Key '%s' already exists. Use Set() instead", key)
	}

	if k, ok := s.refs[parent]; ok {
		s.pushup(k)

		if k == 0 {
			s.stack = append([]interface{}{value}, s.stack)
		} else {
			s.stack = append(s.stack[:k], append([]interface{}{value}, s.stack[k:]...)...)
		}

		s.refs[key] = k
		s.backrefs[k] = key
	} else {
		return errors.Errorf("Key '%s' does not exists", parent)
	}

	return nil
}

// InsertAfter inserts value after parent
func (s *Referenced) InsertAfter(parent string, key string, value interface{}) error {
	if _, ok := s.refs[key]; ok {
		return errors.Errorf("Key '%s' already exists. Use Set() instead", key)
	}

	if k, ok := s.refs[parent]; ok {
		k = k + 1
		s.pushup(k)

		s.stack = append(s.stack[:k], append([]interface{}{value}, s.stack[k:]...)...)
		s.refs[key] = k
		s.backrefs[k] = key
	} else {
		return errors.Errorf("Key '%s' does not exists", parent)
	}

	return nil
}

// Pop value from the top of the stack
func (s *Referenced) Pop() interface{} {
	k := len(s.stack) - 1
	value := s.stack[k]

	s.Unset(s.backrefs[k])
	return value
}

// Unset deletes value from stack by key
func (s *Referenced) Unset(key string) error {
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
func (s *Referenced) Stack() []interface{} {
	return s.stack
}

// Value returns value by key
func (s *Referenced) Value(key string) interface{} {
	if i, ok := s.refs[key]; ok {
		return s.stack[i]
	}

	return nil
}

// Map retruns stack as map
func (s *Referenced) Map() map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range s.stack {
		m[s.backrefs[k]] = v
	}

	return m
}

// Contains returns true if stack contains a value
func (s *Referenced) Contains(value interface{}) bool {
	for _, v := range s.stack {
		if v == value {
			return true
		}
	}

	return false
}

// Has returns true if stack contains a key
func (s *Referenced) Has(key string) bool {
	if _, ok := s.refs[key]; ok {
		return true
	}

	return false
}

// Clear clears the stack
func (s *Referenced) Clear() error {
	s.stack = make([]interface{}, 0)
	s.refs = make(map[string]int)
	s.backrefs = make(map[int]string)
	return nil
}

// Populate the stack with provided values
func (s *Referenced) Populate(data map[string]interface{}) {
	for k, v := range data {
		s.Append(k, v)
	}
}

func (s *Referenced) pushup(key int) {
	newbackrefs := make(map[int]string)
	for k := range s.stack {
		if k >= key {
			newbackrefs[k+1] = s.backrefs[k]
			s.refs[s.backrefs[k]] = k + 1
		}
	}
	s.backrefs = newbackrefs
}

func (s *Referenced) pushdown(key int) {
	newbackrefs := make(map[int]string)
	for k := range s.stack {
		if k >= key {
			newbackrefs[k-1] = s.backrefs[k]
			s.refs[s.backrefs[k]] = k - 1
		}
	}
	s.backrefs = newbackrefs
}

// NewReferenced creates a new referenced stack
func NewReferenced(data map[string]interface{}) *Referenced {
	s := &Referenced{
		refs:     make(map[string]int),
		stack:    make([]interface{}, 0),
		backrefs: make(map[int]string),
	}

	if data != nil {
		s.Populate(data)
	}

	return s
}
