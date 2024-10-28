package controller

import (
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/session"
	"github.com/noxyicm/wsf/view"
)

const (
	// TYPEHelperFlashMessenger represents FlashMessenger action helper
	TYPEHelperFlashMessenger = "flashMessenger"
)

func init() {
	RegisterHelper(TYPEHelperFlashMessenger, NewFlashMessengerHelper)
}

// FlashMessenger is a action helper that handles persistent messeges
type FlashMessenger struct {
	name      string
	Namespace string
	View      view.Interface
}

// Name returns helper name
func (h *FlashMessenger) Name() string {
	return h.name
}

// Init the helper
func (h *FlashMessenger) Init(options map[string]interface{}) error {
	return nil
}

// PreDispatch do dispatch preparations
func (h *FlashMessenger) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch do dispatch aftermath
func (h *FlashMessenger) PostDispatch(ctx context.Context) error {
	return nil
}

// Add adds a message to flash messenger
func (h *FlashMessenger) Add(ctx context.Context, message string, namespace string) error {
	if namespace == "" {
		namespace = h.Namespace
	}

	var ses session.Interface
	if sesInterface := ctx.Value(context.SessionKey); sesInterface != nil {
		ses = sesInterface.(session.Interface)
	} else {
		return errors.New("Unable to get session")
	}

	var m map[string][]string
	if !ses.Has(h.name) {
		m = make(map[string][]string)
		ses.Set(h.name, m)
	} else {
		m = ses.Get(h.name).(map[string][]string)
	}

	if v, ok := m[namespace]; ok && v != nil {
		m[namespace] = append(v, message)
		return nil
	}

	m[namespace] = make([]string, 0)
	m[namespace] = append(m[namespace], message)

	return nil
}

// Has reports wether a specific namespace has messages
func (h *FlashMessenger) Has(ctx context.Context, namespace string) bool {
	if namespace == "" {
		namespace = h.Namespace
	}

	var ses session.Interface
	if sesInterface := ctx.Value(context.SessionKey); sesInterface != nil {
		ses = sesInterface.(session.Interface)
	} else {
		return false
	}

	if !ses.Has(h.name) {
		return false
	}

	if m := ses.Get(h.name); m != nil {
		if v, ok := m.(map[string][]string); ok {
			if _, ok := v[namespace]; ok {
				return true
			}
		}
	}

	return false
}

// Messages returns messages from a specific namespace
func (h *FlashMessenger) Messages(ctx context.Context, namespace string) []string {
	if namespace == "" {
		namespace = h.Namespace
	}

	s := make([]string, 0)
	var ses session.Interface
	if sesInterface := ctx.Value(context.SessionKey); sesInterface != nil {
		ses = sesInterface.(session.Interface)
	} else {
		return s
	}

	if !ses.Has(h.name) {
		return s
	}

	if m := ses.Get(h.name); m != nil {
		if v, ok := m.(map[string][]string); ok {
			if vs, ok := v[namespace]; ok {
				return vs
			}
		}
	}

	return s
}

// ClearMessages clears all messages from the previous request & current namespace
func (h *FlashMessenger) ClearMessages(ctx context.Context, namespace string) bool {
	if namespace == "" {
		namespace = h.Namespace
	}

	if h.Has(ctx, namespace) {
		var ses session.Interface
		if sesInterface := ctx.Value(context.SessionKey); sesInterface != nil {
			ses = sesInterface.(session.Interface)
		} else {
			return false
		}

		ses.Set(h.name, make(map[string][]string))
		return true
	}

	return false
}

// NewFlashMessengerHelper creates new FlashMessenger action helper
func NewFlashMessengerHelper() (HelperInterface, error) {
	return &FlashMessenger{
		name:      "FlashMessenger",
		Namespace: "default",
	}, nil
}
