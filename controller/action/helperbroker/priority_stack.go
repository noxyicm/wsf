package helperbroker

import (
	"wsf/controller/action/helper"
	"wsf/errors"
)

// PriorityStack holds helpers stack sorted by priority
type PriorityStack struct {
	helpers             []helper.Interface
	helpersByPriority   map[int]int
	helpersByName       map[string]int
	nextDefaultPriority int
}

// Push adds helper into stack
func (s *PriorityStack) Push(h helper.Interface) *PriorityStack {
	s.Set(s.NextFreeHigherPriority(s.nextDefaultPriority), h)
	return s
}

// Has returns true if stack contains helper by name
func (s *PriorityStack) Has(name string) bool {
	_, ok := s.helpersByName[name]
	return ok
}

// HasPriority returns true if stack contains helper by priority
func (s *PriorityStack) HasPriority(priority int) bool {
	_, ok := s.helpersByPriority[priority]
	return ok
}

// Get returns helper by name
func (s *PriorityStack) Get(name string) helper.Interface {
	if v, ok := s.helpersByName[name]; ok {
		return s.helpers[v]
	}

	return nil
}

// Set sets helper to priority
func (s *PriorityStack) Set(priority int, h helper.Interface) error {
	if s.Has(h.Name()) {
		s.Unset(h.Name())
	}

	if s.HasPriority(priority) {
		priority = s.NextFreeHigherPriority(priority)
		// need warning here
	}

	key := len(s.helpers)
	for pr, k := range s.helpersByPriority {
		if priority > pr && key < k {
			key = k
		}
	}

	if len(s.helpers) == 0 {
		s.helpers = append(s.helpers, h)
	} else if key == 0 {
		s.helpers = append([]helper.Interface{h}, s.helpers...)
	} else {
		s.helpers = append(s.helpers[:key], append([]helper.Interface{h}, s.helpers[key:]...)...)
	}

	s.helpersByPriority[priority] = key
	s.helpersByName[h.Name()] = key

	if nextFreeDefault := s.NextFreeHigherPriority(s.nextDefaultPriority); priority == nextFreeDefault {
		s.nextDefaultPriority = nextFreeDefault
	}

	return nil
}

// Unset removes helper from stack by name
func (s *PriorityStack) Unset(name string) error {
	if !s.Has(name) {
		return errors.Errorf("A helper with name '%s' does not exist", name)
	}

	key := s.helpersByName[name]
	priority, err := s.FindPriority(s.helpers[key])
	if err != nil {
		return err
	}

	s.helpers = append(s.helpers[:key-1], s.helpers[key:]...)
	delete(s.helpersByName, name)
	delete(s.helpersByPriority, priority)
	return nil
}

// UnsetPriority removes helper from stack by priority
func (s *PriorityStack) UnsetPriority(priority int) error {
	if !s.HasPriority(priority) {
		return errors.Errorf("A helper with priority '%v' does not exist", priority)
	}

	key := s.helpersByPriority[priority]
	name := s.helpers[key].Name()

	s.helpers = append(s.helpers[:key-1], s.helpers[key:]...)
	delete(s.helpersByName, name)
	delete(s.helpersByPriority, priority)
	return nil
}

// FindPriority returns helper priority by its name
func (s *PriorityStack) FindPriority(h helper.Interface) (int, error) {
	for k, v := range s.helpersByPriority {
		if s.helpers[v] == h {
			return k, nil
		}
	}

	return 0, errors.Errorf("A helper with name '%s' does not exist", h.Name())
}

// NextFreeHigherPriority finds the next free higher priority. If an index is given, it will
// find the next free highest priority after it.
func (s *PriorityStack) NextFreeHigherPriority(priority int) int {
	priorities := make([]int, len(s.helpersByPriority))
	i := 0
	for k := range s.helpersByPriority {
		priorities[i] = k
		i++
	}

	if i == 0 {
		return priority
	}

	found := false
	for !found {
		for _, v := range priorities {
			if v != priority {
				found = true
				break
			}

			priority++
		}
	}

	return priority
}

// NextFreeLowerPriority finds the next free lower priority.  If an index is given, it will
// find the next free lower priority before it.
func (s *PriorityStack) NextFreeLowerPriority(priority int) int {
	priorities := make([]int, len(s.helpersByPriority))
	i := 0
	for k := range s.helpersByPriority {
		priorities[i] = k
		i++
	}

	if i == 0 {
		return priority
	}

	found := false
	for !found {
		for _, v := range priorities {
			if v != priority {
				found = true
				break
			}

			priority--
		}
	}

	return priority
}

// HighestPriority returns the highest priority
func (s *PriorityStack) HighestPriority() int {
	max := -int(^uint(0)>>1) - 1
	for k := range s.helpersByPriority {
		if k > max {
			max = k
		}
	}

	if max == -int(^uint(0)>>1)-1 {
		return s.nextDefaultPriority
	}

	return max
}

// LowestPriority returns the lowest priority
func (s *PriorityStack) LowestPriority() int {
	min := int(^uint(0) >> 1)
	for k := range s.helpersByPriority {
		if k < min {
			min = k
		}
	}

	if min == int(^uint(0)>>1) {
		return s.nextDefaultPriority
	}

	return min
}

// Helpers returns the helpers stack
func (s *PriorityStack) Helpers() []helper.Interface {
	return s.helpers
}

// HelpersByName returns the helpers referenced by name
func (s *PriorityStack) HelpersByName() map[string]helper.Interface {
	m := make(map[string]helper.Interface)
	for _, v := range s.helpers {
		m[v.Name()] = v
	}

	return m
}

// Clear clears the stack
func (s *PriorityStack) Clear() {
	s.helpers = make([]helper.Interface, 0)
	s.helpersByName = make(map[string]int)
	s.helpersByPriority = make(map[int]int)
	s.nextDefaultPriority = 0
}

// NewPriorityStack creates new PriorityStack
func NewPriorityStack() (*PriorityStack, error) {
	return &PriorityStack{
		helpers:             make([]helper.Interface, 0),
		helpersByPriority:   make(map[int]int),
		helpersByName:       make(map[string]int),
		nextDefaultPriority: 0,
	}, nil
}
