package view

import (
	"html/template"

	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/utils"
)

const (
	// TYPEViewHelperHeadScript is the name of view helper
	TYPEViewHelperHeadScript = "headScript"
)

func init() {
	RegisterHelper(TYPEViewHelperHeadScript, NewHeadScript)
}

// HeadScript is a paginator view halper
type HeadScript struct {
	name    string
	tag     string
	scripts []map[string]string
	view    Interface
}

// Name returns helper name
func (h *HeadScript) Name() string {
	return h.name
}

// Init the helper
func (h *HeadScript) Init(vi Interface, options map[string]interface{}) error {
	h.SetView(vi)

	if err := vi.AddTemplateFunc(h.name, h.RenderContent); err != nil {
		return errors.Wrap(err, "Unable to add function to template")
	}

	return nil
}

// Setup the helper
func (h *HeadScript) Setup() error {
	return nil
}

// SetView sets view
func (h *HeadScript) SetView(vi Interface) error {
	h.view = vi
	return nil
}

// Render renders helper content
func (h *HeadScript) Render() error {
	return nil
}

func (h *HeadScript) RenderContent() template.HTML {
	rendered := ``
	for _, script := range h.scripts {
		rendered += `<script type="` + script["type"] + `"`
		if script["src"] != "" {
			rendered += ` src="` + script["src"] + `"`
		}
		text := script["script"]
		delete(script, "type")
		delete(script, "src")
		delete(script, "script")
		for attr, attrValue := range script {
			rendered += ` ` + attr + `="` + attrValue + `"`
		}
		rendered += `>`
		if text != "" {
			//rendered += `//<![CDATA[` + text + `//]]`
			rendered += text
		}
		rendered += `</script>`
	}

	return template.HTML(rendered)
}

func (h *HeadScript) AppendFile(src string, attrs map[string]string) error {
	s := map[string]string{
		"type":   "text/javascript",
		"src":    src,
		"script": "",
	}
	h.scripts = append(h.scripts, utils.MapSSMerge(attrs, s))
	return nil
}

func (h *HeadScript) PrependFile(src string, attrs map[string]string) error {
	s := map[string]string{
		"type":   "text/javascript",
		"src":    src,
		"script": "",
	}
	h.scripts = append([]map[string]string{utils.MapSSMerge(attrs, s)}, h.scripts...)
	return nil
}

func (h *HeadScript) AppendScript(script string, attrs map[string]string) error {
	s := map[string]string{
		"type":   "text/javascript",
		"src":    "",
		"script": script,
	}
	h.scripts = append(h.scripts, utils.MapSSMerge(attrs, s))
	return nil
}

func (h *HeadScript) PrependScript(script string, attrs map[string]string) error {
	s := map[string]string{
		"type":   "text/javascript",
		"src":    "",
		"script": script,
	}
	h.scripts = append([]map[string]string{utils.MapSSMerge(attrs, s)}, h.scripts...)
	return nil
}

// NewHeadScript creates a new HeadScript view halper
func NewHeadScript() (HelperInterface, error) {
	return &HeadScript{
		name:    "HeadScript",
		tag:     "script",
		scripts: make([]map[string]string, 0),
	}, nil
}
