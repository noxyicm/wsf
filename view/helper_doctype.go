package view

import (
	"strings"

	"github.com/noxyicm/wsf/errors"
)

// DocType constants
const (
	XHTML11            = "XHTML11"
	XHTML1Strict       = "XHTML1_STRICT"
	XHTML1Transitional = "XHTML1_TRANSITIONAL"
	XHTML1Frameset     = "XHTML1_FRAMESET"
	XHTML1Rdfa         = "XHTML1_RDFA"
	XHTML1Rdfa11       = "XHTML1_RDFA11"
	XHTMLBasic1        = "XHTML_BASIC1"
	XHTML5             = "XHTML5"
	HTML4Strict        = "HTML4_STRICT"
	HTML4Loose         = "HTML4_LOOSE"
	HTML4Frameset      = "HTML4_FRAMESET"
	HTML5              = "HTML5"
	CUSTOMXHTML        = "CUSTOM_XHTML"
	CUSTOM             = "CUSTOM"

	// TYPEHelperDoctype is the name of view helper
	TYPEHelperDoctype = "doctype"
)

func init() {
	RegisterHelper(TYPEHelperDoctype, NewDoctypeHelper)
}

// Doctype is a DocType view helper
type Doctype struct {
	name     string
	registry map[string]interface{}
	view     Interface
}

// Name returns helper name
func (h *Doctype) Name() string {
	return h.name
}

// Init the helper
func (h *Doctype) Init(vi Interface, options map[string]interface{}) error {
	h.SetView(vi)
	return nil
}

// Setup the helper
func (h *Doctype) Setup() error {
	return nil
}

// SetView sets view
func (h *Doctype) SetView(vi Interface) error {
	h.view = vi
	return nil
}

// Render renders helper content
func (h *Doctype) Render() error {
	return nil
}

// SetDoctype sets a DocType
func (h *Doctype) SetDoctype(doctype string) error {
	switch doctype {
	case XHTML11, XHTML1Strict, XHTML1Transitional, XHTML1Frameset, XHTMLBasic1, XHTML1Rdfa, XHTML1Rdfa11, XHTML5, HTML4Strict, HTML4Loose, HTML4Frameset, HTML5:
		h.registry["doctype"] = doctype

	default:
		if doctype[0:9] != "<!DOCTYPE" {
			return errors.Errorf("The specified doctype is malformed")
		}

		typ := CUSTOM
		if strings.Contains(doctype, "xhtml") {
			typ = CUSTOMXHTML
		}

		h.registry["doctype"] = doctype
		h.registry["doctypes"].(map[string]string)[typ] = doctype
	}

	return nil
}

// GetDoctype returns DocType
func (h *Doctype) GetDoctype() string {
	return h.registry["doctypes"].(map[string]string)[h.registry["doctype"].(string)]
}

// IsXhtml returns true if doctype is XHTML
func (h *Doctype) IsXhtml() bool {
	return strings.Contains(h.registry["doctype"].(string), "xhtml")
}

// IsStrict returns true if doctype is strict html
func (h *Doctype) IsStrict() bool {
	switch h.registry["doctype"] {
	case XHTML1Strict, XHTML11, HTML4Strict:
		return true
	}

	return false
}

// IsHTML5 returns true if doctype is HTML5
func (h *Doctype) IsHTML5() bool {
	return strings.Contains(h.GetDoctype(), "<!DOCTYPE html>")
}

// IsRDFA returns true if doctype is RDFA
func (h *Doctype) IsRDFA() bool {
	return strings.Contains(h.registry["doctype"].(string), "rdfa")
}

// NewDoctypeHelper creates a new Doctype view helper
func NewDoctypeHelper() (HelperInterface, error) {
	doctypes := map[string]string{
		XHTML11:            "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.1//EN\" \"http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd\">",
		XHTML1Strict:       "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\">",
		XHTML1Transitional: "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">",
		XHTML1Frameset:     "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Frameset//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-frameset.dtd\">",
		XHTML1Rdfa:         "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.0//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd\">",
		XHTML1Rdfa11:       "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.1//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-2.dtd\">",
		XHTMLBasic1:        "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML Basic 1.0//EN\" \"http://www.w3.org/TR/xhtml-basic/xhtml-basic10.dtd\">",
		XHTML5:             "<!DOCTYPE html>",
		HTML4Strict:        "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01//EN\" \"http://www.w3.org/TR/html4/strict.dtd\">",
		HTML4Loose:         "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\" \"http://www.w3.org/TR/html4/loose.dtd\">",
		HTML4Frameset:      "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Frameset//EN\" \"http://www.w3.org/TR/html4/frameset.dtd\">",
		HTML5:              "<!DOCTYPE html>",
	}

	return &Doctype{
		name:     "Doctype",
		registry: map[string]interface{}{"doctypes": doctypes, "doctype": HTML5},
	}, nil
}
