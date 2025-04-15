package controller

import (
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"
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
func (h *HelperBroker) SetHelper(priority int, hlp HelperInterface, replace bool, options map[string]interface{}) error {
	err := h.stack.Set(priority, hlp, replace)
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

// GetHelper returns a registered action helper by name
func (h *HelperBroker) GetHelper(name string) (HelperInterface, error) {
	name = h.normalizeHelperName(name)
	if h.stack.Has(name) {
		return h.stack.Get(name), nil
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
	defer h.dispatchRecover(ctx)

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
	defer h.dispatchRecover(ctx)

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

func (h *HelperBroker) dispatchRecover(ctx context.Context) {
	if r := recover(); r != nil {
		switch er := r.(type) {
		case error:
			ctx.AddError(errors.Wrap(er, "Unxpected error equired"))

		default:
			ctx.AddError(errors.Errorf("Unxpected error equired: %v", er))
		}
	}
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
