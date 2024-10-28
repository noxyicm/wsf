package controller

import (
	"wsf/context"
	"wsf/errors"
)

var broker *HelperBroker

// HelperBroker stores and dispatches action helpers
type HelperBroker struct {
	stack *HelpersStack
}

// AddHelper pushs helper into stack
func (h *HelperBroker) AddHelper(hlp HelperInterface) error {
	h.stack.Push(hlp)
	return hlp.Init(nil)
}

// SetHelper sets helper into stack with priority
func (h *HelperBroker) SetHelper(priority int, hlp HelperInterface, options map[string]interface{}) error {
	err := h.stack.Set(priority, hlp)
	if err != nil {
		return err
	}

	return hlp.Init(options)
}

// HasHelper returns true if action helper by name is registered
func (h *HelperBroker) HasHelper(name string) bool {
	name = h.normalizeHelperName(name)
	return h.stack.Has(name)
}

// GetExistingHelper returns a registered action helper by name
func (h *HelperBroker) GetExistingHelper(name string) (HelperInterface, error) {
	name = h.normalizeHelperName(name)
	if h.stack.Has(name) {
		return h.stack.Get(name), nil
	}

	return nil, errors.Errorf("Helper by name %s does not exists", name)
}

// GetHelper returns a registered action helper by name and initializes if needed
func (h *HelperBroker) GetHelper(name string) (HelperInterface, error) {
	name = h.normalizeHelperName(name)
	if h.stack.Has(name) {
		return h.stack.Get(name), nil
	}

	if hlpr, err := NewHelper(name); err == nil {
		if err := h.AddHelper(hlpr); err != nil {
			return nil, err
		}

		return hlpr, nil
	}

	return nil, errors.Errorf("Helper by name %s does not exists", name)
}

// RemoveHelper removes a registered action helper by its name
func (h *HelperBroker) RemoveHelper(name string) error {
	name = h.normalizeHelperName(name)
	if h.stack.Has(name) {
		return h.stack.Unset(name)
	}

	return errors.Errorf("Helper by name %s does not exists", name)
}

// ResetHelpers clears the stack
func (h *HelperBroker) ResetHelpers() {
	h.stack.Clear()
}

// NotifyPreDispatch notifyes action helpers of preDispatch state
func (h *HelperBroker) NotifyPreDispatch(ctx context.Context) error {
	for _, v := range h.stack.Helpers() {
		err := v.PreDispatch(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// NotifyPostDispatch notifyes action helpers of postDispatch state
func (h *HelperBroker) NotifyPostDispatch(ctx context.Context) error {
	for _, v := range h.stack.Helpers() {
		err := v.PostDispatch(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *HelperBroker) normalizeHelperName(name string) string {
	return name
}

// NewHelperBroker creates new HelperBroker
func NewHelperBroker() (*HelperBroker, error) {
	hb := &HelperBroker{}
	stack, err := NewHelpersStack()
	if err != nil {
		return nil, err
	}

	hb.stack = stack
	broker = hb
	return hb, nil
}
