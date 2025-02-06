package view

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
	"github.com/noxyicm/wsf/utils"
	"github.com/noxyicm/wsf/view/helper/placeholder"
)

// DocType constants
const (
	// TYPEHelperHeadMeta is the name of view helper
	TYPEHelperHeadMeta = "headmeta"
)

func init() {
	RegisterHelper(TYPEHelperHeadMeta, NewHeadMetaHelper)
}

// HeadMeta view helper
type HeadMeta struct {
	placeholder.Standalone

	name         string
	typeKeys     []string
	requiredKeys []string
	modifierKeys []string
	view         Interface
}

// Name returns helper name
func (h *HeadMeta) Name() string {
	return h.name
}

// Init the helper
func (h *HeadMeta) Init(vi Interface, options map[string]interface{}) error {
	h.SetView(vi)

	if err := vi.AddTemplateFunc(h.name, h.RenderContent); err != nil {
		return errors.Wrap(err, "Unable to add function to template")
	}

	return nil
}

// Setup the helper
func (h *HeadMeta) Setup() error {
	return nil
}

// SetView sets view
func (h *HeadMeta) SetView(vi Interface) error {
	h.view = vi
	return nil
}

// Render renders helper content
func (h *HeadMeta) Render() error {
	return nil
}

func (h *HeadMeta) RenderContent(data map[string]interface{}) template.HTML {
	rendered, err := h.view.Render(data, "views_helpers_"+h.name, h.view.GetOptions().SegmentContentKey)
	if err != nil {
		log.Notice(fmt.Sprintf("View_Helper_HeadMeta error render content: %v", err), nil)
		return ""
	}

	return template.HTML(rendered)
}

// Set sets a value
func (h *HeadMeta) Set(value interface{}) error {
	if !h.isValid(value) {
		return errors.New("Invalid value passed to set; please use setMeta()")
	}

	for _, item := range h.Container().GetStack() {
		if item.(*HeadMetaData).Type == value.(*HeadMetaData).Type && item.(*HeadMetaData).TypeValue == value.(*HeadMetaData).TypeValue {
			h.Container().Unset(value.(*HeadMetaData).TypeValue)
		}
	}

	return h.Container().Append(value.(*HeadMetaData).TypeValue, value)
}

// SetCharset sets a charset
func (h *HeadMeta) SetCharset(charset string) error {
	item := h.createData("charset", charset, "", map[string]string{})
	h.Set(item)
	return nil
}

// Get returns HeadMetaData struct from container
func (h *HeadMeta) Get(key string) *HeadMetaData {
	if v := h.Container().Get(key); v != nil {
		return v.(*HeadMetaData)
	}

	return nil
}

// AppendName appends name meta tag
func (h *HeadMeta) AppendName(typeValue string, content string, modifiers map[string]string) {
	item := h.createData("name", typeValue, content, modifiers)
	h.Container().Append(typeValue, item)
}

// PrependName prepends name meta
func (h *HeadMeta) PrependName(typeValue string, content string, modifiers map[string]string) {
	item := h.createData("name", typeValue, content, modifiers)
	h.Container().Prepend(typeValue, item)
}

// SetName sets a single name meta
func (h *HeadMeta) SetName(typeValue string, content string, modifiers map[string]string) {
	item := h.createData("name", typeValue, content, modifiers)
	h.Set(item)
}

// AppendHTTPEquiv appends http-equiv meta
func (h *HeadMeta) AppendHTTPEquiv(typeValue string, content string, modifiers map[string]string) {
	item := h.createData("http-equiv", typeValue, content, modifiers)
	h.Container().Append(typeValue, item)
}

// PrependHTTPEquiv prepends http-equiv meta
func (h *HeadMeta) PrependHTTPEquiv(typeValue string, content string, modifiers map[string]string) {
	item := h.createData("http-equiv", typeValue, content, modifiers)
	h.Container().Prepend(typeValue, item)
}

// SetHTTPEquiv sets a single http-equiv meta
func (h *HeadMeta) SetHTTPEquiv(typeValue string, content string, modifiers map[string]string) {
	item := h.createData("http-equiv", typeValue, content, modifiers)
	h.Set(item)
}

// AppendProperty appends property meta
func (h *HeadMeta) AppendProperty(typeValue string, content string, modifiers map[string]string) {
	item := h.createData("property", typeValue, content, modifiers)
	h.Container().Append(typeValue, item)
}

// PrependProperty prepends property meta
func (h *HeadMeta) PrependProperty(typeValue string, content string, modifiers map[string]string) {
	item := h.createData("property", typeValue, content, modifiers)
	h.Container().Prepend(typeValue, item)
}

// SetProperty sets a single property meta
func (h *HeadMeta) SetProperty(typeValue string, content string, modifiers map[string]string) {
	item := h.createData("property", typeValue, content, modifiers)
	h.Set(item)
}

func (h *HeadMeta) createData(typ, typeValue, content string, modifiers map[string]string) *HeadMetaData {
	return &HeadMetaData{
		Type:      typ,
		TypeValue: typeValue,
		Content:   content,
		Modifiers: modifiers,
	}
}

func (h *HeadMeta) isValid(item interface{}) bool {
	var metaItem *HeadMetaData
	if v, ok := item.(*HeadMetaData); ok {
		metaItem = v
	} else {
		return false
	}

	isHTML5 := false
	if h.view != nil && h.view.Helper("Doctype") != nil && h.view.Helper("Doctype").(*Doctype).IsHTML5() {
		isHTML5 = true
	}

	if metaItem.Content == "" && (!isHTML5 || (!isHTML5 && metaItem.Type != "charset")) {
		return false
	}

	// <meta property= ... /> is only supported with doctype RDFa
	if h.view != nil && h.view.Helper("Doctype") != nil && !h.view.Helper("Doctype").(*Doctype).IsRDFA() && metaItem.Type == "property" {
		return false
	}

	return true
}

func (h *HeadMeta) itemToString(item *HeadMetaData) (string, error) {
	if !utils.InSSlice(item.Type, h.typeKeys) {
		return "", errors.Errorf("Invalid type '%s' provided for meta", item.Type)
	}

	typ := item.Type
	modifiersString := ""
	for key, value := range item.Modifiers {
		if h.view != nil && h.view.Helper("Doctype") != nil && h.view.Helper("Doctype").(*Doctype).IsHTML5() && key == "scheme" {
			return "", errors.Errorf("Invalid modifier '%s' provided; not supported by HTML5", key)
		}

		if !utils.InSSlice(key, h.modifierKeys) {
			continue
		}

		modifiersString += key + "=\"" + h.Escape(value) + "\" "
	}

	tpl := "<meta %s=\"%s\" content=\"%s\" %s/>"
	if h.view != nil && h.view.Helper("Doctype") != nil {
		if h.view.Helper("Doctype").(*Doctype).IsHTML5() && typ == "charset" {
			if h.view.Helper("Doctype").(*Doctype).IsXhtml() {
				tpl = "<meta %s=\"%s\"/>"
			} else {
				tpl = "<meta %s=\"%s\">"
			}
		} else if h.view.Helper("Doctype").(*Doctype).IsXhtml() {
			tpl = "<meta %s=\"%s\" content=\"%s\" %s/>"
		} else {
			tpl = "<meta %s=\"%s\" content=\"%s\" %s>"
		}
	}

	meta := fmt.Sprintf(tpl, typ, h.Escape(item.Type), h.Escape(item.Content), modifiersString)

	if v, ok := item.Modifiers["conditional"]; ok {
		v = strings.ReplaceAll(v, " ", "")
		if v == "!IE" {
			meta = "<!-->" + meta + "<!--"
		}

		meta = "<!--[if " + h.Escape(v) + "]>" + meta + "<![endif]-->"
	}

	return meta, nil
}

func (h *HeadMeta) toString() string {
	items := []string{}
	//h.Container().Sort()
	for _, item := range h.Container().GetStack() {
		stringified, err := h.itemToString(item.(*HeadMetaData))
		if err != nil {
			log.Notice(err.Error(), nil)
		}

		items = append(items, stringified)
	}

	return h.Container().Indent() + strings.Join(items, h.Escape(h.Container().Separator())+h.Container().Indent())
}

func (h *HeadMeta) normalizeType(typ string) (string, error) {
	switch typ {
	case "Name":
		return "name", nil

	case "HttpEquiv":
		return "http-equiv", nil

	case "Property":
		return "property", nil
	}

	return "", errors.Errorf("Invalid type '%s' passed to normalizeType", typ)
}

// NewHeadMetaHelper creates a new HeadMeta view helper
func NewHeadMetaHelper() (HelperInterface, error) {
	hm := &HeadMeta{
		name:         "HeadMeta",
		typeKeys:     []string{"name", "http-equiv", "charset", "property"},
		requiredKeys: []string{"content"},
		modifierKeys: []string{"lang", "scheme"},
	}
	hm.Standalone.RegKey = "WSFViewHelperHeadMeta"
	hm.Registry = placeholder.GetRegistry()
	hm.SetContainer(hm.Registry.GetContainer(hm.RegKey))
	return hm, nil
}

// HeadMetaData type
type HeadMetaData struct {
	Type      string
	TypeValue string
	Content   string
	Modifiers map[string]string
}
