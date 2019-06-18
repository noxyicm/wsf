package stack

import "wsf/errors"

// Prioritised stack
type Prioritised struct {
	stack      []interface{}
	priorities map[int]int
}

// Push value into stack
func (s *Prioritised) Push(priority int, value interface{}) error {
	key := len(s.stack)
	for k, pr := range s.priorities {
		if priority > pr && key < k {
			key = k
		}
	}

	if len(s.stack) == 0 {
		s.stack = append(s.stack, value)
	} else if key == 0 {
		s.stack = append([]interface{}{value}, s.stack...)
	} else {
		s.stack = append(s.stack[:key], append([]interface{}{value}, s.stack[key:]...)...)
	}

	s.priorities[key] = priority
	return nil
}

// Pop value from the top of the stack
func (s *Prioritised) Pop() interface{} {
	k := len(s.stack) - 1
	value := s.stack[k]

	s.Unset(k)
	return value
}

// Unset deletes value from stack by key
func (s *Prioritised) Unset(key int) error {
	if _, ok := s.priorities[key]; ok {
		s.stack = append(s.stack[:key], s.stack[key:]...)
		delete(s.priorities, key)
		s.pushdown(key)
	} else {
		return errors.Errorf("Key '%d' does not exists", key)
	}

	return nil
}

// Stack returns all values
func (s *Prioritised) Stack() []interface{} {
	return s.stack
}

// Value returns value by key
func (s *Prioritised) Value(key int) interface{} {
	if _, ok := s.priorities[key]; ok {
		return s.stack[key]
	}

	return nil
}

// Contains returns true if stack contains a value
func (s *Prioritised) Contains(value interface{}) bool {
	for _, v := range s.stack {
		if v == value {
			return true
		}
	}

	return false
}

// Has returns true if stack contains a key
func (s *Prioritised) Has(key int) bool {
	if _, ok := s.priorities[key]; ok {
		return true
	}

	return false
}

// Clear clears the stack
func (s *Prioritised) Clear() error {
	s.stack = make([]interface{}, 0)
	s.priorities = make(map[int]int)
	return nil
}

// Len returns length of stack
func (s *Prioritised) Len() int {
	return len(s.stack)
}

func (s *Prioritised) pushup(key int) {
	newpriorities := make(map[int]int)
	for k := range s.stack {
		if k >= key {
			newpriorities[k+1] = s.priorities[k]
		}
	}
	s.priorities = newpriorities
}

func (s *Prioritised) pushdown(key int) {
	newpriorities := make(map[int]int)
	for k := range s.stack {
		if k >= key {
			newpriorities[k-1] = s.priorities[k]
		}
	}
	s.priorities = newpriorities
}

// NewPrioritised creates a new prioritised stack
func NewPrioritised() *Prioritised {
	return &Prioritised{
		stack:      make([]interface{}, 0),
		priorities: make(map[int]int),
	}
}
