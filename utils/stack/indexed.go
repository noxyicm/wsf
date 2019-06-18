package stack

import "wsf/errors"

// Indexed stack
type Indexed struct {
	stack []interface{}
}

// Set values as stack
func (s *Indexed) Set(key int, value interface{}) error {
	if len(s.stack) > key {
		s.stack[key] = value
		return nil
	}

	return errors.Errorf("Key '%d' does not exists", key)
}

// Append value to stack
func (s *Indexed) Append(value interface{}) error {
	s.stack = append(s.stack, value)
	return nil
}

// Prepend value to stack
func (s *Indexed) Prepend(value interface{}) error {
	s.stack = append([]interface{}{value}, s.stack)
	return nil
}

// InsertBefore inserts value before parent
func (s *Indexed) InsertBefore(parent int, value interface{}) error {
	if len(s.stack) > parent {
		if parent == 0 {
			s.stack = append([]interface{}{value}, s.stack)
		} else {
			s.stack = append(s.stack[:parent], append([]interface{}{value}, s.stack[parent:]...)...)
		}
	} else {
		return errors.Errorf("Key '%d' does not exists", parent)
	}

	return nil
}

// InsertAfter inserts value after parent
func (s *Indexed) InsertAfter(parent int, value interface{}) error {
	if len(s.stack) > parent {
		parent = parent + 1
		s.stack = append(s.stack[:parent], append([]interface{}{value}, s.stack[parent:]...)...)
	} else {
		return errors.Errorf("Key '%d' does not exists", parent)
	}

	return nil
}

// Pop value from the top of the stack
func (s *Indexed) Pop() interface{} {
	k := len(s.stack) - 1
	value := s.stack[k]
	s.Unset(k)
	return value
}

// Unset deletes value from stack by key
func (s *Indexed) Unset(key int) error {
	if len(s.stack) > key {
		s.stack = append(s.stack[:key], s.stack[key:]...)
	} else {
		return errors.Errorf("Key '%d' does not exists", key)
	}

	return nil
}

// Stack returns all values
func (s *Indexed) Stack() []interface{} {
	return s.stack
}

// Value returns value by key
func (s *Indexed) Value(key int) interface{} {
	if len(s.stack) > key {
		return s.stack[key]
	}

	return nil
}

// Contains returns true if stack contains a value
func (s *Indexed) Contains(value interface{}) bool {
	for _, v := range s.stack {
		if v == value {
			return true
		}
	}

	return false
}

// Has returns true if stack contains a key
func (s *Indexed) Has(key int) bool {
	if len(s.stack) > key {
		return true
	}

	return false
}

// Clear clears the stack
func (s *Indexed) Clear() error {
	s.stack = make([]interface{}, 0)
	return nil
}

// NewIndexed creates a new indexed stack
func NewIndexed() *Indexed {
	return &Indexed{
		stack: make([]interface{}, 0),
	}
}
