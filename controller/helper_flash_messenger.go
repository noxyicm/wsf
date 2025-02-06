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
	enabled   bool
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
	if session.Created() {
		h.enabled = true
	}
	return nil
}

// PreDispatch do dispatch preparations
func (h *FlashMessenger) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch do dispatch aftermath
func (h *FlashMessenger) PostDispatch(ctx context.Context) error {
	defer h.dispatchRecover(ctx)

	infoFlashMessages := h.Messages(ctx, "info")
	errorFlashMessages := h.Messages(ctx, "error")
	successFlashMessages := h.Messages(ctx, "success")
	flashMessages := map[string]interface{}{
		"showFlashMessages": false,
		"hasError":          false,
		"hasInfo":           false,
		"hasSuccesses":      false,
		"messages": map[string][]string{
			"info":    infoFlashMessages,
			"error":   errorFlashMessages,
			"success": successFlashMessages,
		},
	}
	if len(infoFlashMessages) > 0 || len(errorFlashMessages) > 0 || len(successFlashMessages) > 0 {
		flashMessages["showFlashMessages"] = true

		if len(infoFlashMessages) > 0 {
			flashMessages["hasInfo"] = true
		}

		if len(errorFlashMessages) > 0 {
			flashMessages["hasError"] = true
		}

		if len(successFlashMessages) > 0 {
			flashMessages["hasSuccesses"] = true
		}
	}
	ctx.SetDataValue("flashMessages", flashMessages)
	return nil
}

// Add adds a message to flash messenger
func (h *FlashMessenger) Add(ctx context.Context, message string, namespace string) error {
	if !h.enabled {
		return nil
	}

	if namespace == "" {
		namespace = h.Namespace
	}

	var ses session.Interface
	if sesInterface := ctx.Value(context.SessionKey); sesInterface != nil {
		ses = sesInterface.(session.Interface)
	} else {
		return errors.New("Unable to get session")
	}

	var m map[string][]*FlashMessage
	if !ses.Has(h.name) {
		m = make(map[string][]*FlashMessage)
		ses.Set(h.name, m)
	} else {
		m = h.getValues(ses)
		ses.Set(h.name, m)
	}

	if v, ok := m[namespace]; ok && v != nil {
		m[namespace] = append(v, &FlashMessage{
			Message: message,
			Expires: 1,
		})

		return nil
	}

	m[namespace] = make([]*FlashMessage, 0)
	m[namespace] = append(m[namespace], &FlashMessage{
		Message: message,
		Expires: 1,
	})

	return nil
}

// Has reports wether a specific namespace has messages
func (h *FlashMessenger) Has(ctx context.Context, namespace string) bool {
	if !h.enabled {
		return false
	}

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

	m := h.getValues(ses)
	if _, ok := m[namespace]; ok {
		return true
	}

	return false
}

// Messages returns messages from a specific namespace
func (h *FlashMessenger) Messages(ctx context.Context, namespace string) []string {
	s := make([]string, 0)
	if !h.enabled {
		return s
	}

	if namespace == "" {
		namespace = h.Namespace
	}

	var ses session.Interface
	if sesInterface := ctx.Value(context.SessionKey); sesInterface != nil {
		ses = sesInterface.(session.Interface)
	} else {
		return s
	}

	if !ses.Has(h.name) {
		return s
	}

	m := h.getValues(ses)
	if v, ok := m[namespace]; ok {
		s := make([]string, len(v))
		mv := make([]*FlashMessage, len(v))
		copy(mv, v)
		for i, vs := range mv {
			s[i] = vs.Message
			if vs.Expires > 0 && !ctx.Response().IsRedirect() {
				vs.Expires--
			}

			if vs.Expires == 0 {
				m[namespace] = h.removeExpired(m[namespace], vs)
			}
		}

		ses.Set(h.name, m)
		return s
	}

	return s
}

// ClearMessages clears all messages from the previous request & current namespace
func (h *FlashMessenger) ClearMessages(ctx context.Context, namespace string) bool {
	if !h.enabled {
		return false
	}

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

		ses.Set(h.name, make(map[string][]*FlashMessage))
		return true
	}

	return false
}

func (h *FlashMessenger) getValues(ses session.Interface) (m map[string][]*FlashMessage) {
	m = make(map[string][]*FlashMessage)
	if !ses.Has(h.name) {
		return m
	}

	if im := ses.Get(h.name); im != nil {
		if iv, ok := im.(map[string]interface{}); ok {
			for key, v := range iv {
				if vs, ok := v.([]*FlashMessage); ok {
					m[key] = vs
				} else if vs, ok := v.([]interface{}); ok {
					m[key] = make([]*FlashMessage, len(vs))
					for i, vss := range vs {
						m[key][i] = NewFlashMessage(vss)
					}
				}
			}
		} else if iv, ok := im.(map[string][]*FlashMessage); ok {
			return iv
		}
	}

	return m
}

func (h *FlashMessenger) removeExpired(s []*FlashMessage, v *FlashMessage) []*FlashMessage {
	i := 0
	var vs *FlashMessage
	found := false
	for i, vs = range s {
		if v == vs {
			found = true
			break
		}
	}

	if !found {
		return s
	}

	if i == 0 {
		s = s[1:]
	} else if i == len(s)-1 {
		s = s[:len(s)-1]
	} else {
		s = append(s[0:i-1], s[i:]...)
	}

	return s
}

func (h *FlashMessenger) dispatchRecover(ctx context.Context) {
	if r := recover(); r != nil {
		switch er := r.(type) {
		case error:
			ctx.AddError(errors.Wrap(er, "Unxpected error equired"))

		default:
			ctx.AddError(errors.Errorf("Unxpected error equired: %v", er))
		}
	}
}

// NewFlashMessengerHelper creates new FlashMessenger action helper
func NewFlashMessengerHelper() (HelperInterface, error) {
	return &FlashMessenger{
		enabled:   false,
		name:      "FlashMessenger",
		Namespace: "default",
	}, nil
}

type FlashMessage struct {
	Message string `json:"message"`
	Expires int    `json:"exipres"`
}

func NewFlashMessage(data interface{}) *FlashMessage {
	fm := &FlashMessage{
		Message: "",
		Expires: 0,
	}

	if d, ok := data.(map[string]interface{}); ok {
		if msg, ok := d["message"]; ok {
			fm.Message = msg.(string)
		}

		if exp, ok := d["exipres"]; ok {
			fm.Expires = int(exp.(float64))
		}
	}

	return fm
}
