package helper

import (
	"fmt"
	"strings"
	"wsf/errors"
	"wsf/log"
	"wsf/utils"
	"wsf/view/helper/placeholder"
)

// HeadMeta view helper
type HeadMeta struct {
	placeholder.Standalone
	typeKeys     []string
	requiredKeys []string
	modifierKeys []string
	view         ViewInterface
}

// SetView sets view
func (hm *HeadMeta) SetView(vi ViewInterface) error {
	hm.view = vi
	return nil
}

// Set sets a value
func (hm *HeadMeta) Set(value interface{}) error {
	if !hm.isValid(value) {
		return errors.New("Invalid value passed to set; please use setMeta()")
	}

	for _, item := range hm.Container().GetStack() {
		if item.(*HeadMetaData).Type == value.(*HeadMetaData).Type && item.(*HeadMetaData).TypeValue == value.(*HeadMetaData).TypeValue {
			hm.Container().Unset(value.(*HeadMetaData).TypeValue)
		}
	}

	return hm.Container().Append(value.(*HeadMetaData).TypeValue, value)
}

// SetCharset sets a charset
func (hm *HeadMeta) SetCharset(charset string) error {
	item := hm.createData("charset", charset, "", map[string]string{})
	hm.Set(item)
	return nil
}

// Get returns HeadMetaData struct from container
func (hm *HeadMeta) Get(key string) *HeadMetaData {
	if v := hm.Container().Get(key); v != nil {
		return v.(*HeadMetaData)
	}

	return nil
}

// AppendName appends name meta tag
func (hm *HeadMeta) AppendName(typeValue string, content string, modifiers map[string]string) {
	item := hm.createData("name", typeValue, content, modifiers)
	hm.Container().Append(typeValue, item)
}

// PrependName prepends name meta
func (hm *HeadMeta) PrependName(typeValue string, content string, modifiers map[string]string) {
	item := hm.createData("name", typeValue, content, modifiers)
	hm.Container().Prepend(typeValue, item)
}

// SetName sets a single name meta
func (hm *HeadMeta) SetName(typeValue string, content string, modifiers map[string]string) {
	item := hm.createData("name", typeValue, content, modifiers)
	hm.Set(item)
}

// AppendHTTPEquiv appends http-equiv meta
func (hm *HeadMeta) AppendHTTPEquiv(typeValue string, content string, modifiers map[string]string) {
	item := hm.createData("http-equiv", typeValue, content, modifiers)
	hm.Container().Append(typeValue, item)
}

// PrependHTTPEquiv prepends http-equiv meta
func (hm *HeadMeta) PrependHTTPEquiv(typeValue string, content string, modifiers map[string]string) {
	item := hm.createData("http-equiv", typeValue, content, modifiers)
	hm.Container().Prepend(typeValue, item)
}

// SetHTTPEquiv sets a single http-equiv meta
func (hm *HeadMeta) SetHTTPEquiv(typeValue string, content string, modifiers map[string]string) {
	item := hm.createData("http-equiv", typeValue, content, modifiers)
	hm.Set(item)
}

// AppendProperty appends property meta
func (hm *HeadMeta) AppendProperty(typeValue string, content string, modifiers map[string]string) {
	item := hm.createData("property", typeValue, content, modifiers)
	hm.Container().Append(typeValue, item)
}

// PrependProperty prepends property meta
func (hm *HeadMeta) PrependProperty(typeValue string, content string, modifiers map[string]string) {
	item := hm.createData("property", typeValue, content, modifiers)
	hm.Container().Prepend(typeValue, item)
}

// SetProperty sets a single property meta
func (hm *HeadMeta) SetProperty(typeValue string, content string, modifiers map[string]string) {
	item := hm.createData("property", typeValue, content, modifiers)
	hm.Set(item)
}

func (hm *HeadMeta) createData(typ, typeValue, content string, modifiers map[string]string) *HeadMetaData {
	return &HeadMetaData{
		Type:      typ,
		TypeValue: typeValue,
		Content:   content,
		Modifiers: modifiers,
	}
}

func (hm *HeadMeta) isValid(item interface{}) bool {
	var metaItem *HeadMetaData
	if v, ok := item.(*HeadMetaData); ok {
		metaItem = v
	} else {
		return false
	}

	isHTML5 := false
	if hm.view != nil && hm.view.Helper("Doctype") != nil && hm.view.Helper("Doctype").(*Doctype).IsHTML5() {
		isHTML5 = true
	}

	if metaItem.Content == "" && (!isHTML5 || (!isHTML5 && metaItem.Type != "charset")) {
		return false
	}

	// <meta property= ... /> is only supported with doctype RDFa
	if hm.view != nil && hm.view.Helper("Doctype") != nil && !hm.view.Helper("Doctype").(*Doctype).IsRDFA() && metaItem.Type == "property" {
		return false
	}

	return true
}

func (hm *HeadMeta) itemToString(item *HeadMetaData) (string, error) {
	if !utils.InSSlice(item.Type, hm.typeKeys) {
		return "", errors.Errorf("Invalid type '%s' provided for meta", item.Type)
	}

	typ := item.Type
	modifiersString := ""
	for key, value := range item.Modifiers {
		if hm.view != nil && hm.view.Helper("Doctype") != nil && hm.view.Helper("Doctype").(*Doctype).IsHTML5() && key == "scheme" {
			return "", errors.Errorf("Invalid modifier '%s' provided; not supported by HTML5", key)
		}

		if !utils.InSSlice(key, hm.modifierKeys) {
			continue
		}

		modifiersString += key + "=\"" + hm.Escape(value) + "\" "
	}

	tpl := "<meta %s=\"%s\" content=\"%s\" %s/>"
	if hm.view != nil && hm.view.Helper("Doctype") != nil {
		if hm.view.Helper("Doctype").(*Doctype).IsHTML5() && typ == "charset" {
			if hm.view.Helper("Doctype").(*Doctype).IsXhtml() {
				tpl = "<meta %s=\"%s\"/>"
			} else {
				tpl = "<meta %s=\"%s\">"
			}
		} else if hm.view.Helper("Doctype").(*Doctype).IsXhtml() {
			tpl = "<meta %s=\"%s\" content=\"%s\" %s/>"
		} else {
			tpl = "<meta %s=\"%s\" content=\"%s\" %s>"
		}
	}

	meta := fmt.Sprintf(tpl, typ, hm.Escape(item.Type), hm.Escape(item.Content), modifiersString)

	if v, ok := item.Modifiers["conditional"]; ok {
		v = strings.ReplaceAll(v, " ", "")
		if v == "!IE" {
			meta = "<!-->" + meta + "<!--"
		}

		meta = "<!--[if " + hm.Escape(v) + "]>" + meta + "<![endif]-->"
	}

	return meta, nil
}

func (hm *HeadMeta) toString() string {
	items := []string{}
	//hm.Container().Sort()
	for _, item := range hm.Container().GetStack() {
		stringified, err := hm.itemToString(item.(*HeadMetaData))
		if err != nil {
			log.Notice(err.Error(), nil)
		}

		items = append(items, stringified)
	}

	return hm.Container().Indent() + strings.Join(items, hm.Escape(hm.Container().Separator())+hm.Container().Indent())
}

func (hm *HeadMeta) normalizeType(typ string) (string, error) {
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

// NewHeadMeta creates a new HeadMeta helper
func NewHeadMeta() (*HeadMeta, error) {
	hm := &HeadMeta{
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
