package controller

import (
	"wsf/config"
	"wsf/context"
	"wsf/errors"
	"wsf/utils"
)

// Public contants
const (
	ContextSwitchTriggerINIT = "TRIGGER_INIT"
	ContextSwitchTriggerPOST = "TRIGGER_POST"
)

const (
	// TYPEHelperContextSwitch represents ContextSwitch action helper
	TYPEHelperContextSwitch = "contextSwitch"
)

func init() {
	RegisterHelper(TYPEHelperContextSwitch, NewContextSwitchHelper)
}

// ContextSwitch is a context switch helper
type ContextSwitch struct {
	name           string
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
	return h.name
}

// Init initializes at start of action controller
// Reset the view script suffix to the original state, or store the
// original state.
func (h *ContextSwitch) Init(options map[string]interface{}) error {
	//cs.SetOptions(options)
	if h.ViewSuffixOrig == "" {
		h.ViewSuffixOrig = h.ViewRenderer().ViewSuffix()
	}

	return nil
}

// PreDispatch satisfys action helper interface
func (h *ContextSwitch) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch satisfys action helper interface
func (h *ContextSwitch) PostDispatch(ctx context.Context) error {
	return nil
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
func (h *ContextSwitch) InitContext(ctx context.Context, format string) error {
	h.CurrentContext = ""
	req := ctx.Request()
	action := req.ActionName()

	// Return if no context switching enabled, or no context switching
	// enabled for this action
	ctxs := h.ActionContexts(ctx, action)
	if len(ctxs) == 0 {
		return nil
	}

	// Return if no context parameter provided
	ctxName := req.ParamString(h.ContextParam)
	if ctxName == "" {
		if format == "" {
			return nil
		}

		ctxName = format
		format = ""
	}

	// Check if context allowed by action controller
	if !h.HasActionContext(ctx, action, ctxName) {
		return nil
	}

	// Return if invalid context parameter provided and no format or invalid
	// format provided
	if !h.HasContext(ctxName) {
		if format == "" || !h.HasContext(format) {
			return nil
		}
	}

	// Use provided format if passed
	if format != "" && h.HasContext(format) {
		ctxName = format
	}

	suffix := h.Suffix(ctxName)
	h.ViewRenderer().SetViewSuffix(suffix)

	if headers, ok := h.Headers(ctxName); ok {
		rsp := ctx.Response()
		for header, content := range headers {
			rsp.SetHeader(header, content)
		}
	}

	if h.AutoDisableLayout() {
		ctx.SetParam(context.LayoutEnabledKey, false)
	}

	h.CurrentContext = ctxName
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
func (h *ContextSwitch) SetSuffix(ctxName, suffix string, prependViewRendererSuffix bool) *ContextSwitch {
	if _, ok := h.Contexts[ctxName]; !ok {
		return h
	}

	if prependViewRendererSuffix {
		if suffix == "" {
			suffix = h.ViewRenderer().ViewSuffix()
		} else {
			suffix = suffix + "." + h.ViewRenderer().ViewSuffix()
		}
	}

	h.Contexts[ctxName].Suffix = suffix
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

// AutoDisableLayout retrieves auto layout disable flag
func (h *ContextSwitch) AutoDisableLayout() bool {
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
func (h *ContextSwitch) AddActionContext(ctx context.Context, action string, ctxName string) error {
	if !h.HasContext(ctxName) {
		return errors.Errorf("Context '%s' does not exist", ctx)
	}

	data := ctx.Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		m[action] = append(m[action], ctxName)
		ctx.SetParam(h.ContextKey, m)
	} else {
		ctx.SetParam(h.ContextKey, map[string][]string{
			action: {
				ctxName,
			},
		})
	}

	return nil
}

// SetActionContext sets a context as available for a given controller action
func (h *ContextSwitch) SetActionContext(ctx context.Context, action string, ctxNames []string) error {
	for _, ctx := range ctxNames {
		if !h.HasContext(ctx) {
			return errors.Errorf("Context '%s' does not exist", ctx)
		}
	}

	data := ctx.Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		m[action] = ctxNames
		ctx.SetParam(h.ContextKey, m)
	} else {
		ctx.SetParam(h.ContextKey, map[string][]string{
			action: ctxNames,
		})
	}

	return nil
}

// AddActionContexts adds multiple action/context pairs at once
func (h *ContextSwitch) AddActionContexts(ctx context.Context, specs map[string][]string) error {
	for action, ctxs := range specs {
		for _, ctxName := range ctxs {
			if err := h.AddActionContext(ctx, action, ctxName); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetActionContexts overwrites and set multiple action contexts at once
func (h *ContextSwitch) SetActionContexts(ctx context.Context, specs map[string][]string) error {
	for action, ctxs := range specs {
		if err := h.SetActionContext(ctx, action, ctxs); err != nil {
			return err
		}
	}

	return nil
}

// HasActionContext does a particular controller action have the given context?
func (h *ContextSwitch) HasActionContext(ctx context.Context, action string, ctxName string) bool {
	if !h.HasContext(ctxName) {
		return false
	}

	data := ctx.Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		if v, ok := m[action]; ok {
			if utils.InSSlice(ctxName, v) {
				return true
			}
		}
	}

	return false
}

// ActionContexts gets contexts for a given action or all actions in the controller
func (h *ContextSwitch) ActionContexts(ctx context.Context, action string) []string {
	data := ctx.Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		if v, ok := m[action]; ok {
			return v
		}
	}

	return make([]string, 0)
}

// AllActionContexts gets contexts for a given action or all actions in the controller
func (h *ContextSwitch) AllActionContexts(ctx context.Context) map[string][]string {
	data := ctx.Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		return m
	}

	return make(map[string][]string)
}

// RemoveActionContext removes context for a given controller action
func (h *ContextSwitch) RemoveActionContext(ctx context.Context, action string, typ string) bool {
	if !h.HasActionContext(ctx, action, typ) {
		return false
	}

	data := ctx.Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		if actionContexts, ok := m[action]; ok {
			if indx, ok := utils.SKey(typ, actionContexts); ok {
				actionContexts = append(actionContexts[:indx], actionContexts[indx+1:]...)
				m[action] = actionContexts
				ctx.SetParam(h.ContextKey, m)
			}
		}
	}

	return false
}

// ClearActionContexts clears all contexts for a given controller action or all actions
func (h *ContextSwitch) ClearActionContexts(ctx context.Context, action string) bool {
	data := ctx.Param(h.ContextKey)
	if m, ok := data.(map[string][]string); ok {
		delete(m, action)
		ctx.SetParam(h.ContextKey, m)
	}

	return false
}

// ViewRenderer retrieves ViewRenderer
func (h *ContextSwitch) ViewRenderer() *ViewRenderer {
	if h.viewRenderer == nil {
		if Instance().HasHelper("viewRenderer") {
			h.viewRenderer = Instance().Helper("viewRenderer").(*ViewRenderer)
			return h.viewRenderer
		}
	}

	return nil
}

// NewContextSwitchHelper creates a new context switch helper
func NewContextSwitchHelper() (HelperInterface, error) {
	cs := &ContextSwitch{
		name:           "ContextSwitch",
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

// SwitchableContext represetns a context that context switch can use
type SwitchableContext struct {
	Name    string
	Suffix  string
	Headers map[string]string
}
