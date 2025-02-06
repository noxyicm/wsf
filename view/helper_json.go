package view

import (
	"encoding/json"

	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
)

const (
	// TYPEViewHelperJSON is the name of view helper
	TYPEViewHelperJSON = "json"
)

func init() {
	RegisterHelper(TYPEViewHelperJSON, NewJSON)
}

// JSON is a view halper for showing session related messages
type JSON struct {
	name string
	view Interface
}

// Name returns helper name
func (h *JSON) Name() string {
	return h.name
}

// Init the helper
func (h *JSON) Init(vi Interface, options map[string]interface{}) error {
	h.SetView(vi)

	if err := vi.AddTemplateFunc(h.name, h.RenderContent); err != nil {
		return errors.Wrap(err, "Unable to add function to template")
	}

	return nil
}

// Setup the helper
func (h *JSON) Setup() error {
	return nil
}

// SetView sets view
func (h *JSON) SetView(vi Interface) error {
	h.view = vi
	return nil
}

// Render renders helper content
func (h *JSON) Render() error {
	return nil
}

func (h *JSON) RenderContent(data map[string]interface{}) string {
	rendered, err := json.Marshal(data)
	if err != nil {
		log.Notice("[View_Helper_JSON] error equired while rendering content: "+err.Error(), nil)
		return ""
	}

	return string(rendered)
}

// NewJSON creates a new JSON view halper
func NewJSON() (HelperInterface, error) {
	return &JSON{
		name: "JSON",
	}, nil
}
