package helper

import (
	"wsf/config"
	"wsf/context"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/session"
	"wsf/utils"
)

// Public contants
const (
	TriggerINIT = "TRIGGER_INIT"
	TriggerPOST = "TRIGGER_POST"
)

const (
	// TYPEContextSwitch represents ContextSwitch action helper
	TYPEContextSwitch = "contextSwitch"
)

func init() {
	Register(TYPEContextSwitch, NewContextSwitchAsHelper)
}

// ContextSwitch is a context switch helper
type ContextSwitch struct {
	controller     ControllerInterface
	Contexts       map[string]*SwitchableContext
	ContextKey     string
	ContextParam   string
	CurrentContext string
	DefaultContext string
	DisableLayout  bool
	SpecialConfig  []string
	Unconfigurable []string
	ViewSuffixOrig string
	viewRenderer   *ViewRenderer
}

// Name satisfys action helper interface
func (h *ContextSwitch) Name() string {
	return "ContextSwitch"
}

// Init initializes at start of action controller
// Reset the view script suffix to the original state, or store the
// original state.
func (h *ContextSwitch) Init(options map[string]interface{}) error {
	if h.ViewSuffixOrig == "" {
		h.ViewSuffixOrig = h.ViewRenderer().ViewSuffix()
	} else {
		h.ViewRenderer().SetViewSuffix(h.ViewSuffixOrig)
	}

	return nil
}

// PreDispatch satisfys action helper interface
func (h *ContextSwitch) PreDispatch(ctx context.Context) error {
	return nil
}

//PostDispatch satisfys action helper interface
func (h *ContextSwitch) PostDispatch(ctx context.Context) error {
	return nil
}

// SetController satisfys action helper interface
func (h *ContextSwitch) SetController(ctrl ControllerInterface) error {
	h.controller = ctrl
	return nil
}

// Controller satisfys action helper interface
func (h *ContextSwitch) Controller() ControllerInterface {
	return h.controller
}

// Request satisfys action helper interface
func (h *ContextSwitch) Request() request.Interface {
	return h.controller.Request()
}

// Response satisfys action helper interface
func (h *ContextSwitch) Response() response.Interface {
	return h.controller.Response()
}

// Session satisfys action helper interface
func (h *ContextSwitch) Session() session.Interface {
	return h.controller.Session()
}

// SetOptions configure struct from options
func (h *ContextSwitch) SetOptions(options config.Config) *ContextSwitch {
	if options.Get("contexts") != nil {
		h.SetContexts(options.GetStringMap("contexts"))
	}

	for _, key := range options.GetKeys() {
		if utils.InSSlice(key, h.Unconfigurable) {
			continue
		}

		switch key {
		case "suffix":
			h.setSuffix(options.GetStringMap(key))
		}
	}

	return h
}

// InitContext initialize context detection and switching
func (h *ContextSwitch) InitContext(format string) error {
	h.CurrentContext = ""
	req := h.Request()
	action := req.ActionName()

	// Return if no context switching enabled, or no context switching
	// enabled for this action
	ctxs := h.ActionContexts(action)
	if len(ctxs) == 0 {
		return nil
	}

	// Return if no context parameter provided
	ctx := req.ParamString(h.ContextParam)
	if ctx == "" {
		if format == "" {
			return nil
		}

		ctx = format
		format = ""
	}

	// Check if context allowed by action controller
	if !h.HasActionContext(action, ctx) {
		return nil
	}

	// Return if invalid context parameter provided and no format or invalid
	// format provided
	if !h.HasContext(ctx) {
		if format == "" || !h.HasContext(format) {
			return nil
		}
	}

	// Use provided format if passed
	if format != "" && h.HasContext(format) {
		ctx = format
	}

	suffix := h.Suffix(ctx)
	h.ViewRenderer().SetViewSuffix(suffix)

	if headers, ok := h.Headers(ctx); ok {
		rsp := h.Response()
		for header, content := range headers {
			rsp.SetHeader(header, content)
		}
	}

	if h.GetAutoDisableLayout() {
		ctrl := h.Controller()
		if ctrl != nil {
			ctrl.Context().SetValue(context.LayoutEnabledKey, false)
			// make some thinking
			// layout struct is single but disable only for this context
			//layout.(*layout.Interface).DisableLayout()
		}
	}

	h.CurrentContext = ctx
	return nil
}

// setSuffix sets suffix from map
func (h *ContextSwitch) setSuffix(spec interface{}) *ContextSwitch {
	switch m := spec.(type) {
	case map[string]interface{}:
		for ctx, suffixInfo := range m {
			switch v := suffixInfo.(type) {
			case string:
				h.SetSuffix(ctx, v, true)

			case map[string]interface{}:
				if suf, ok := v["suffix"]; ok {
					suffix := suf.(string)
					prependViewRendererSuffix := true

					if prep, ok := v["prependviewrenderersuffix"]; ok {
						prependViewRendererSuffix = prep.(bool)
					}

					h.SetSuffix(ctx, suffix, prependViewRendererSuffix)
				}

			case []string:
				if len(v) > 0 {
					h.SetSuffix(ctx, v[0], true)
				}
			}
		}

	case []string:
		for _, suffix := range m {
			h.SetSuffix(suffix, suffix, true)
		}
	}

	return h
}

// SetSuffix customize view script suffix to use when switching context.
// Passing an empty suffix value to the setters disables the view script
// suffix change.
func (h *ContextSwitch) SetSuffix(ctx, suffix string, prependViewRendererSuffix bool) *ContextSwitch {
	if _, ok := h.Contexts[ctx]; !ok {
		return h
	}

	if prependViewRendererSuffix {
		if suffix == "" {
			suffix = h.ViewRenderer().ViewSuffix()
		} else {
			suffix = suffix + "." + h.ViewRenderer().ViewSuffix()
		}
	}

	h.Contexts[ctx].Suffix = suffix
	return h
}

// Suffix retrieve suffix for given context type
func (h *ContextSwitch) Suffix(ctx string) string {
	if v, ok := h.Contexts[ctx]; ok {
		return v.Suffix
	}

	return ""
}

// HasContext checks if the given context exist
func (h *ContextSwitch) HasContext(ctx string) bool {
	if _, ok := h.Contexts[ctx]; ok {
		return true
	}

	return false
}

// AddHeader adds header to context
func (h *ContextSwitch) AddHeader(ctx, header, content string) error {
	if !h.HasContext(ctx) {
		return errors.Errorf("Context '%s' does not exist", ctx)
	}

	if _, ok := h.Contexts[ctx].Headers[header]; ok {
		return errors.Errorf("Cannot add '%s' header to context '%s': already exists", header, ctx)
	}

	h.Contexts[ctx].Headers[header] = content
	return nil
}

// SetHeader customize response header to use when switching context
func (h *ContextSwitch) SetHeader(ctx, header, content string) error {
	if !h.HasContext(ctx) {
		return errors.Errorf("Context '%s' does not exist", ctx)
	}

	h.Contexts[ctx].Headers[header] = content
	return nil
}

// AddHeaders adds multiple headers at once for a given context
func (h *ContextSwitch) AddHeaders(ctx string, headers map[string]string) error {
	for header, content := range headers {
		if err := h.AddHeader(ctx, header, content); err != nil {
			return err
		}
	}

	return nil
}

// setHeaders sets headers from context => headers pairs
func (h *ContextSwitch) setHeaders(options map[string]interface{}) error {
	for ctx, headers := range options {
		switch m := headers.(type) {
		case map[string]interface{}:
			h.ClearHeaders(ctx)
			for header, content := range m {
				if err := h.SetHeader(ctx, header, content.(string)); err != nil {
					return err
				}
			}

		case map[string]string:
			if err := h.SetHeaders(ctx, m); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetHeaders sets multiple headers at once for a given context
func (h *ContextSwitch) SetHeaders(ctx string, headers map[string]string) error {
	h.ClearHeaders(ctx)
	for header, content := range headers {
		if err := h.SetHeader(ctx, header, content); err != nil {
			return err
		}
	}

	return nil
}

// Header retrieves context header
func (h *ContextSwitch) Header(ctx, header string) (string, bool) {
	if !h.HasContext(ctx) {
		return "", false
	}

	if v, ok := h.Contexts[ctx].Headers[header]; ok {
		return v, ok
	}

	return "", false
}

// Headers retrieves context headers
func (h *ContextSwitch) Headers(ctx string) (map[string]string, bool) {
	if !h.HasContext(ctx) {
		return nil, false
	}

	return h.Contexts[ctx].Headers, true
}

// RemoveHeader removes a single header from a context
func (h *ContextSwitch) RemoveHeader(ctx, header string) bool {
	if !h.HasContext(ctx) {
		return false
	}

	if _, ok := h.Contexts[ctx].Headers[header]; ok {
		delete(h.Contexts[ctx].Headers, header)
		return true
	}

	return false
}

// ClearHeaders clears all headers for a given context
func (h *ContextSwitch) ClearHeaders(ctx string) bool {
	if !h.HasContext(ctx) {
		return false
	}

	h.Contexts[ctx].Headers = make(map[string]string)
	return true
}

// SetContextParam sets name of parameter to use when determining context format
func (h *ContextSwitch) SetContextParam(name string) error {
	h.ContextParam = name
	return nil
}

// GetContextParam returns context format request parameter name
func (h *ContextSwitch) GetContextParam() string {
	return h.ContextParam
}

// SetDefaultContext indicates default context to use when no context format provided
func (h *ContextSwitch) SetDefaultContext(ctx string) error {
	if !h.HasContext(ctx) {
		return errors.Errorf("Cannot set default context; invalid context type '%s'", ctx)
	}

	h.DefaultContext = ctx
	return nil
}

// GetDefaultContext returns default context
func (h *ContextSwitch) GetDefaultContext() string {
	return h.DefaultContext
}

// SetAutoDisableLayout sets flag indicating if layout should be disabled
func (h *ContextSwitch) SetAutoDisableLayout(flag bool) error {
	h.DisableLayout = flag
	return nil
}

// GetAutoDisableLayout retrieves auto layout disable flag
func (h *ContextSwitch) GetAutoDisableLayout() bool {
	return h.DisableLayout
}

// AddContext adds new context
func (h *ContextSwitch) AddContext(ctx string, spec map[string]interface{}) error {
	if h.HasContext(ctx) {
		return errors.Errorf("Cannot add context '%s'; already exists", ctx)
	}

	h.Contexts[ctx] = &SwitchableContext{
		Name:    ctx,
		Suffix:  "",
		Headers: make(map[string]string),
	}

	if v, ok := spec["suffix"]; ok {
		h.Contexts[ctx].Suffix = v.(string)
	}

	if v, ok := spec["headers"]; ok {
		if m, ok := v.(map[string]string); ok {
			h.Contexts[ctx].Headers = m
		}
	}

	return nil
}

// SetContext overwrites existing context
func (h *ContextSwitch) SetContext(ctx string, spec map[string]interface{}) error {
	h.RemoveContext(ctx)
	return h.AddContext(ctx, spec)
}

// AddContexts adds multiple contexts
func (h *ContextSwitch) AddContexts(ctxs map[string]interface{}) error {
	for ctx, spec := range ctxs {
		switch m := spec.(type) {
		case map[string]interface{}:
			if err := h.AddContext(ctx, m); err != nil {
				return err
			}

		default:
			if err := h.AddContext(ctx, map[string]interface{}{}); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetContexts sets multiple contexts, after first removing all
func (h *ContextSwitch) SetContexts(ctxs map[string]interface{}) error {
	h.ClearContexts()
	return h.AddContexts(ctxs)
}

// Context retrieves context specification
func (h *ContextSwitch) Context(ctx string) *SwitchableContext {
	if h.HasContext(ctx) {
		return h.Contexts[ctx]
	}

	return nil
}

// AllContexts retrieves context definitions
func (h *ContextSwitch) AllContexts() map[string]*SwitchableContext {
	return h.Contexts
}

// RemoveContext removes a context
func (h *ContextSwitch) RemoveContext(ctx string) bool {
	if h.HasContext(ctx) {
		delete(h.Contexts, ctx)
		return true
	}

	return false
}

// ClearContexts removes all contexts
func (h *ContextSwitch) ClearContexts() {
	h.Contexts = make(map[string]*SwitchableContext)
}

// GetCurrentContext returns current context, if any
func (h *ContextSwitch) GetCurrentContext() string {
	return h.CurrentContext
}

// AddActionContext adds one or more contexts to an action
func (h *ContextSwitch) AddActionContext(action, ctx string) error {
	if !h.HasContext(ctx) {
		return errors.Errorf("Context '%s' does not exist", ctx)
	}

	ctrl := h.Controller()
	if ctrl == nil {
		return nil
	}

	data := ctrl.Context().Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		m[action] = append(m[action], ctx)
		ctrl.Context().SetParam(h.ContextKey, m)
	} else {
		ctrl.Context().SetParam(h.ContextKey, map[string][]string{
			action: []string{
				ctx,
			},
		})
	}

	return nil
}

// SetActionContext sets a context as available for a given controller action
func (h *ContextSwitch) SetActionContext(action string, ctxs []string) error {
	for _, ctx := range ctxs {
		if !h.HasContext(ctx) {
			return errors.Errorf("Context '%s' does not exist", ctx)
		}
	}

	ctrl := h.Controller()
	if ctrl == nil {
		return nil
	}

	data := ctrl.Context().Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		m[action] = ctxs
		ctrl.Context().SetParam(h.ContextKey, m)
	} else {
		ctrl.Context().SetParam(h.ContextKey, map[string][]string{
			action: ctxs,
		})
	}

	return nil
}

// AddActionContexts adds multiple action/context pairs at once
func (h *ContextSwitch) AddActionContexts(specs map[string][]string) error {
	for action, ctxs := range specs {
		for _, ctx := range ctxs {
			if err := h.AddActionContext(action, ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetActionContexts overwrites and set multiple action contexts at once
func (h *ContextSwitch) SetActionContexts(specs map[string][]string) error {
	for action, ctxs := range specs {
		if err := h.SetActionContext(action, ctxs); err != nil {
			return err
		}
	}

	return nil
}

// HasActionContext does a particular controller action have the given context?
func (h *ContextSwitch) HasActionContext(action, ctx string) bool {
	if !h.HasContext(ctx) {
		return false
	}

	ctrl := h.Controller()
	if ctrl == nil {
		return false
	}

	data := ctrl.Context().Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		if v, ok := m[action]; ok {
			if utils.InSSlice(ctx, v) {
				return true
			}
		}
	}

	return false
}

// ActionContexts gets contexts for a given action or all actions in the controller
func (h *ContextSwitch) ActionContexts(action string) []string {
	ctrl := h.Controller()
	if ctrl == nil {
		return make([]string, 0)
	}

	data := ctrl.Context().Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		if v, ok := m[action]; ok {
			return v
		}
	}

	return make([]string, 0)
}

// AllActionContexts gets contexts for a given action or all actions in the controller
func (h *ContextSwitch) AllActionContexts() map[string][]string {
	ctrl := h.Controller()
	if ctrl == nil {
		return make(map[string][]string)
	}

	data := ctrl.Context().Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		return m
	}

	return make(map[string][]string)
}

// RemoveActionContext removes context for a given controller action
func (h *ContextSwitch) RemoveActionContext(action, ctx string) bool {
	if !h.HasActionContext(action, ctx) {
		return false
	}

	ctrl := h.Controller()
	if ctrl == nil {
		return false
	}

	data := ctrl.Context().Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		if actionContexts, ok := m[action]; ok {
			if indx, ok := utils.SKey(ctx, actionContexts); ok {
				actionContexts = append(actionContexts[:indx], actionContexts[indx+1:]...)
				m[action] = actionContexts
				ctrl.Context().SetParam(h.ContextKey, m)
			}
		}
	}

	return false
}

// ClearActionContexts clears all contexts for a given controller action or all actions
func (h *ContextSwitch) ClearActionContexts(action string) bool {
	ctrl := h.Controller()
	if ctrl == nil {
		return false
	}

	data := ctrl.Context().Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		delete(m, action)
		ctrl.Context().SetParam(h.ContextKey, m)
	}

	return false
}

// ViewRenderer retrieves ViewRenderer
func (h *ContextSwitch) ViewRenderer() *ViewRenderer {
	if h.viewRenderer == nil {
		ctrl := h.Controller()
		if ctrl == nil {
			return nil
		}

		if ctrl.HasHelper("viewRenderer") {
			h.viewRenderer = ctrl.Helper("viewRenderer").(*ViewRenderer)
			return h.viewRenderer
		}
	}

	return nil
}

// NewContextSwitch creates a new context switch helper
func NewContextSwitch(options config.Config) (*ContextSwitch, error) {
	cs := &ContextSwitch{
		Contexts:       make(map[string]*SwitchableContext),
		ContextKey:     "contexts",
		ContextParam:   "format",
		DefaultContext: "html",
		DisableLayout:  false,
		SpecialConfig: []string{
			"setSuffix",
			"setHeaders",
			"setCallbacks",
		},
		Unconfigurable: []string{
			"setOptions",
			"setConfig",
			"setHeader",
			"setCallback",
			"setContext",
			"setActionContext",
			"setActionContexts",
		},
	}

	cs.SetOptions(options)

	if len(cs.Contexts) == 0 {
		cs.AddContexts(map[string]interface{}{
			"json": map[string]interface{}{
				"suffix": "json",
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
			"xml": map[string]interface{}{
				"suffix": "xml",
				"headers": map[string]string{
					"Content-Type": "application/xml",
				},
			},
		})
	}

	if err := cs.Init(nil); err != nil {
		return nil, errors.Wrap(err, "Unable to create ContextSwitch helper")
	}

	return cs, nil
}

// NewContextSwitchAsHelper satisfys helper constructor
func NewContextSwitchAsHelper() (Interface, error) {
	return NewContextSwitch(nil)
}

// SwitchableContext represetns a context that context switch can use
type SwitchableContext struct {
	Name    string
	Suffix  string
	Headers map[string]string
}
